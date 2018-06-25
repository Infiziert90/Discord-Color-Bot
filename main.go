package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"math/rand"
	"strings"
	"image"
	"image/draw"
	"log"
	"image/png"
	"image/color"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	BotToken string `yaml:"BotToken"`
	InviteLink string `yaml:"InviteLink"`
	AutoKickOnServer map[string]string `yaml:"AutoKickOnServer"`
	SpamChannel string `yaml:"SpamChannel"`
	Admins map[string]string `yaml:"Admins"`
	Colors map[string]int `yaml:"Colors"`

}

type Roles struct {
	ID   string
	Name string
}

var (
	config Config
	RoleNames []string
	CreatedRoles = map[string]map[string]Roles{}
	FirstTime = true
	HelpText = `Help for Color-Bot
<<PrintColors   https://nayu.moe/colors
<<NewColor   "Assign a random color to the current user"
<<NewColor ColorName   "Assign the specified color to the current user"
<<PreviewColor ColorName   "Post a preview image of the color"`
)
func init() {
	createConfig(&config)

	if config.BotToken == "YOUR_BOT_TOKEN" {
		panic("Default BotToken, pls change the settings in config.yaml.")
	} else if config.SpamChannel == "Channel_ID" {
		panic("Default SpamChannel, pls change the settings in config.yaml.")
	}
}

func createConfig(conf *Config) {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

// Main
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	for key := range config.Colors {
		RoleNames = append(RoleNames, key)
	}

	discord, err := discordgo.New("Bot " + config.BotToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	discord.AddHandler(OnReady)
	discord.AddHandler(OnMessage)
	discord.AddHandler(OnMemberJoin)
	discord.AddHandler(MemberChunkRequest)

	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func OnReady(session *discordgo.Session, Ready *discordgo.Ready) {
	session.UpdateStatus(0, "Perfect Color!")

	if FirstTime {
		for _, Guild := range Ready.Guilds {
			LoadRoles(session, Guild.ID)
			CreateNewRoles(session, Guild.ID)
		}
		fmt.Printf("Done.\n")
		FirstTime = false
	}
}

func OnMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Author.ID == session.State.User.ID {
		return
	}

	if !strings.HasPrefix(msg.Content, "<<") {
		return
	}

	Channel, err := session.State.Channel(msg.ChannelID)
	if err != nil {
		fmt.Printf("Can't get the channel.\n")
		fmt.Printf("Error:\n%s", err)
		return
	}

	if CheckChannel(Channel) {
		if strings.HasPrefix(msg.Content, "<<NewColor") {
			err := RemoveColorFromMember(session, Channel.GuildID, msg.Author.ID)
			if err {
				return
			}

			SplitContent := strings.Split(msg.Content, " ")
			if len(SplitContent) == 1 {
				UpdateMemberColorRandom(session, Channel.GuildID, msg.Author.ID)
			} else if len(SplitContent) == 2 {
				if _, ok := config.Colors[SplitContent[1]]; ok {
					UpdateMemberColor(session, Channel.GuildID, msg.Author.ID, SplitContent[1])
				} else {
					UpdateMemberColorRandom(session, Channel.GuildID, msg.Author.ID)
					SendMessageAndDeleteAfterTime(session, msg.ChannelID, "Color not found, pls use <<PrintColors.")
				}
			} else {
				UpdateMemberColorRandom(session, Channel.GuildID, msg.Author.ID)
				SendMessageAndDeleteAfterTime(session, msg.ChannelID, "Too many arguments, pls use <<Help.")
			}
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<PreviewColor") {
			SplitContent := strings.Split(msg.Content, " ")
			if len(SplitContent) == 2 {
				if _, ok := config.Colors[SplitContent[1]]; ok {
					Embed := PreviewRole(session, SplitContent[1])
					SendEmbedAndDeleteAfterTime(session, msg.ChannelID, Embed)
				} else {
					SendMessageAndDeleteAfterTime(session, msg.ChannelID, "Color not found, pls use <<PrintColors.")
				}
			} else {
				SendMessageAndDeleteAfterTime(session, msg.ChannelID, "Too many arguments, pls use <<Help.")
			}
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<PrintColors") {
			SendMessageAndDeleteAfterTime(session, msg.ChannelID, "https://nayu.moe/colors")
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<Help") {
			SendMessageAndDeleteAfterTime(session, msg.ChannelID, HelpText)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
	}

	if CheckAdmin(msg.Author.ID) {
		if strings.HasPrefix(msg.Content, "<<NewServer") {
			JoinedNewGuild(session, Channel.GuildID)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<AddColorToAllMember") {
			AddColorToAllMember(session, Channel.GuildID)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<RemoveAllColors") {
			RemoveAllColors(session, Channel.GuildID)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<ReloadColors") {
			config = Config{}
			createConfig(&config)
			CreateNewRoles(session, Channel.GuildID)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
	}
}

func OnMemberJoin(session *discordgo.Session, Member *discordgo.GuildMemberAdd) {
	UpdateMemberColorRandom(session, Member.GuildID, Member.User.ID)
	if _, ok := config.AutoKickOnServer[Member.GuildID]; ok {
		go KickMemberAfterTime(session, Member.GuildID, Member.User.ID)
	}
}

func CheckAdmin(UserID string) (bool) {
	if _, ok := config.Admins[UserID]; ok {
		return true
	}
	return false
}

func CheckChannel(Channel *discordgo.Channel) (bool) {
	if Channel.GuildID == "221919789017202688" {
		if Channel.ID == "300947822956773376" {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

func MemberChunkRequest(session *discordgo.Session, event *discordgo.GuildMembersChunk) {
	for _, Member := range event.Members {
		UpdateMemberColorRandom(session, event.GuildID, Member.User.ID)
	}
	fmt.Printf("Updated all members.\n")
}

func LoadRoles(session *discordgo.Session, GuildID string) {
	GuildRoles, err := session.GuildRoles(GuildID)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		panic("Can't find the server.")
	}

	// Initialise nested map with GuildID as key
	CreatedRoles[GuildID] = map[string]Roles{}
	for _, Role := range GuildRoles {
		if _, ok := config.Colors[Role.Name]; ok {
			CreatedRoles[GuildID][Role.Name] = Roles{Role.ID, Role.Name}
			CreatedRoles[GuildID][Role.ID]   = Roles{Role.ID, Role.Name}
		}
	}
}

func JoinedNewGuild(session *discordgo.Session, GuildID string) {
	// Initialise nested map with GuildID as key
	CreatedRoles[GuildID] = map[string]Roles{}
	fmt.Printf("Joined a new server: %s\n", GuildID)
	CreateAllRoles(session, GuildID)
}

func AddColorToAllMember(session *discordgo.Session, GuildID string) {
	fmt.Printf("Updating all member with new a color.\n")
	session.RequestGuildMembers(GuildID, "", 0)
}

func RemoveAllColors(session *discordgo.Session, GuildID string) {
	GuildRoles, err := session.GuildRoles(GuildID)
	if err != nil {
		panic("Can't find the server.\n")
	}

	for _, Role := range GuildRoles {
		if _, ok := config.Colors[Role.Name]; ok {
			session.GuildRoleDelete(GuildID, Role.ID)
		}
	}
}

func UpdateMemberColor(s *discordgo.Session, GuildID, MemberID, RoleName string) {
	s.GuildMemberRoleAdd(GuildID, MemberID, CreatedRoles[GuildID][RoleName].ID)
}

func UpdateMemberColorRandom(s *discordgo.Session, GuildID, MemberID string) {
	key := rand.Intn(len(config.Colors))
	s.GuildMemberRoleAdd(GuildID, MemberID, CreatedRoles[GuildID][RoleNames[key]].ID)
}

func CreateColorRole(session *discordgo.Session, GuildID, Name string, Color int) {
	role, err := session.GuildRoleCreate(GuildID)
	if err != nil {
		panic("Wrong Permissions: Can't create a role.")
	}

	fmt.Printf("Name: %s     Int: %d \n", Name, Color)
	Role, _ := session.GuildRoleEdit(GuildID, role.ID, Name, Color, false, 0, false)
	CreatedRoles[GuildID][Role.Name] = Roles{Role.ID, Role.Name}
	CreatedRoles[GuildID][Role.ID]   = Roles{Role.ID, Role.Name}
}

func CreateNewRoles(session *discordgo.Session, GuildID string) {
	for Name, Color := range config.Colors {
		if _, ok := CreatedRoles[GuildID][Name]; !ok {
			CreateColorRole(session, GuildID, Name, Color)
		}
	}
}

func CreateAllRoles(session *discordgo.Session, GuildID string) {
	for Name, Color := range config.Colors {
		CreateColorRole(session, GuildID, Name, Color)
	}
}

func RemoveColorFromMember(session *discordgo.Session, GuildID, MemberID string) (bool) {
	Member, err := session.GuildMember(GuildID, MemberID)
	if err != nil {
		fmt.Printf("Can't get the guild.\n")
		fmt.Printf("Error:\n%s", err)
		return true
	}

	for _, RoleID := range Member.Roles {
		if _, ok := CreatedRoles[GuildID][RoleID]; ok {
			session.GuildMemberRoleRemove(GuildID, MemberID, RoleID)
		}
	}
	return false
}

func PreviewRole(session *discordgo.Session, RoleName string) discordgo.MessageEmbed {
	CreateImageWithColor(config.Colors[RoleName], RoleName)
	Embed := CreateImageEmbed(session, RoleName)

	return Embed
}

// Create Preview Image
func CreateImageWithColor(ColorInt int, ColorName string) {
	size := image.Rect(0, 0, 200, 100)
	rgbaImage := image.NewRGBA(size)
	red := uint8((ColorInt >> 16) & 0xff)
	green := uint8((ColorInt >> 8) & 0xff)
	blue := uint8(ColorInt & 0xff)
	c := color.RGBA{R:red, G:green, B:blue, A:255}
	draw.Draw(rgbaImage, rgbaImage.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)

	f, err := os.Create(ColorName + ".png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, rgbaImage); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func CreateImageEmbed(session *discordgo.Session, ColorName string) discordgo.MessageEmbed {
	Embed := discordgo.MessageEmbed{Title: ColorName, Color: config.Colors[ColorName]}
	FileReader, _ := os.Open(ColorName + ".png")
	Msg, err := session.ChannelFileSend(config.SpamChannel, ColorName + ".png", FileReader)
	if err != nil {
		log.Fatal(err)
		return Embed
	}
	Image := discordgo.MessageEmbedImage{URL: Msg.Attachments[0].URL, Height: 100, Width: 200}
	Embed.Image = &Image

	return Embed
}

func KickMemberAfterTime(session *discordgo.Session, GuildID, MemberID string) {
	time.Sleep(30 * time.Minute)

	Member, err := session.GuildMember(GuildID, MemberID)
	if err != nil {
		fmt.Printf("Member already leaved.\n")
		return
	}

	for _, RoleID := range Member.Roles {
		if _, ok := config.Colors[CreatedRoles[GuildID][RoleID].Name]; ok {
			continue
		} else {
			return
		}
	}

	err = session.GuildMemberDeleteWithReason(GuildID, MemberID, "Not enough roles after 30min.")
	if err != nil {
		panic("Wrong Permissions: Can't kick a member.")
	}

	PrivateChannel, err := session.UserChannelCreate(MemberID)
	if err != nil {
		fmt.Printf("Can't send the message.\n")
		fmt.Printf("Error:\n%s", err)
		return
	}

	session.ChannelMessageSend(PrivateChannel.ID, "You got kicked from the server. Please read the welcome channel.\n" + config.InviteLink)
}

func DeleteMessageAfterTime(session *discordgo.Session, Message *discordgo.Message, Time time.Duration) {
	time.Sleep(Time * time.Minute)
	session.ChannelMessageDelete(Message.ChannelID, Message.ID)
}

func SendMessageAndDeleteAfterTime(session *discordgo.Session, ChannelID, Content string) {
	Message, err := session.ChannelMessageSend(ChannelID, Content)
	if err != nil {
		fmt.Printf("Can't send the message.\n")
		fmt.Printf("Error:\n%s", err)
		return
	}

	go DeleteMessageAfterTime(session, Message, 5)
}

func SendEmbedAndDeleteAfterTime(session *discordgo.Session, ChannelID string, Embed discordgo.MessageEmbed) {
	Message, err := session.ChannelMessageSendEmbed(ChannelID, &Embed)
	if err != nil {
		fmt.Printf("Can't send embed.\n")
		fmt.Printf("Error:\n%s", err)
		return
	}

	go DeleteMessageAfterTime(session, Message, 5)
}
