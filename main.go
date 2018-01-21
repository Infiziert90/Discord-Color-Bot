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
	"bytes"
	_ "image/png"
	"image"
	"github.com/lucasb-eyer/go-colorful"
	"image/draw"
	"log"
	"image/png"
	"strconv"
)

// Settings
var AdminIDs = map[string]string{
	"134750562062303232": "Infi",
	"YOUR_ID":            "YOUR_NAME",
}

var BotToken = "YOUR_BOT_TOKEN"               // Your bot token
var InviteLink = "https://discord.gg/V5vaWwr" // Your invite link
var AutoKick = true                           // if true: member will get kicked after 30min (README.md for more info)
var SpamChannel = "Channel_ID"                // Channel for color image spam

// Stop
var RoleList = map[string]int{
	"SpringGreen4":    0x008B45,
	"LightSlateBlue":  0x8470FF,
	"CadetBlue1":      0x98F5FF,
	"AquaMarine":      0x7FFFD4,
	"Chocolate":       0xD2691E,
	"DarkGreen":       0x006400,
	"DarkOrange":      0xFF8C00,
	"LightSalmon4":    0x8B5742,
	"HotPink":         0xFF69B4,
	"Wheat":           0xF5DEB3,
	"LightGoldenrod":  0xEEDD82,
	"Azure3":          0xC1CDCD,
	"Cyan":            0x00FFFF,
	"Firebrick1":      0xFF3030,
	"Tomato":          0xFF6347,
	"Orange":          0xFFA500,
	"Orchid1":         0xFF83FA,
	"DarkGoldenrod1":  0xFFB90F,
	"Yellow2":         0xEEEE00,
	"MediumTurquoise": 0x48D1CC,
	"Aquamarine3":     0x66CDAA,
	"Burlywood3":      0xCDAA7D,
	"Khaki3":          0xCDC673,
	"LightBlue":       0x7289DA,
	"AstolfoHair":     0xFED5DB,
	"YuzuHair":        0xF7E3C0,
	"ZeonRed":         0xC22F50,
	"NatsumeHair":     0xE67E95,
	"HoloHair":        0xD58138,
	"ChthollyBlue":    0x4C82C2,
	"ChthollyRed":     0xE2455A,
	"Gold":            0xFFD700,
	"DarkSeaGreen":    0x8FBC8F,
	"RemHair":         0xAFD7FC,
	"RamHair":         0xF5A2B4,
}

var RoleNames []string

var RoleNewList = map[string]int{

}

type Roles struct {
	ID   string
	Name string
}

var CreatedRoles = map[string]map[string]Roles{}   // Key = ColorName
var CreatedRolesID = map[string]map[string]Roles{} // Key = RoleID

var FirstTime = true
var HelpText = `Help for Color-Bot
<<PrintColors   "Prints a list of all colors"
<<NewColor   "Assign a random color to the current user"
<<NewColor ColorName   "Assign the specified color to the current user"
<<PreviewColor ColorName   "Assign the specified color to the bot"`

// Create Preview Image
func CreateImageWithColor(ColorInt int, ColorName string) {
	DecimalColor := int64(ColorInt)
	size := image.Rect(0, 0, 200, 100)
	rgbaImage := image.NewRGBA(size)
	c, _ := colorful.Hex(fmt.Sprintf("#%s", strconv.FormatInt(DecimalColor, 16)))
	draw.Draw(rgbaImage, rgbaImage.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)

	f, err := os.Create(fmt.Sprintf("%s.png", ColorName))
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
	Embed := discordgo.MessageEmbed{Title: ColorName, Color: RoleList[ColorName]}
	FileReader, _ := os.Open(fmt.Sprintf("%s.png", ColorName))
	Msg, err := session.ChannelFileSend(SpamChannel, fmt.Sprintf("%s.png", ColorName), FileReader)
	if err != nil {
		log.Fatal(err)
		return Embed
	}
	Image := discordgo.MessageEmbedImage{URL: Msg.Attachments[0].URL, Height: 100, Width: 200}
	Embed.Image = &Image

	return Embed
}

func init() {
	if BotToken == "YOUR_BOT_TOKEN" {
		panic("Default BotToken, pls change the settings in main.go and rebuild.")
	} else if SpamChannel == "Channel_ID" {
		panic("Default SpamChannel, pls change the settings in main.go and rebuild.")
	}
}

// Main
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	for key := range RoleList {
		RoleNames = append(RoleNames, key)
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", BotToken))
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
	if FirstTime {
		session.UpdateStatus(0, "Perfect Color!")

		for _, Guild := range Ready.Guilds {
			loadRoles(session, Guild.ID)
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
				if _, ok := RoleList[SplitContent[1]]; ok {
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
				if _, ok := RoleList[SplitContent[1]]; ok {
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
			var buffer bytes.Buffer
			for key := range RoleList {
				buffer.WriteString(key + "\n")
			}
			buffer.WriteString("Use `<<PreviewColor ColorName` for a preview.")

			SendMessageAndDeleteAfterTime(session, msg.ChannelID, buffer.String())
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
		} else if strings.HasPrefix(msg.Content, "<<AddColorAllMember") {
			AddAllMembersColor(session, Channel.GuildID)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if strings.HasPrefix(msg.Content, "<<RemoveAllColors") {
			RemoveAllColors(session, Channel.GuildID)
			session.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
	}
}

func OnMemberJoin(session *discordgo.Session, Member *discordgo.GuildMemberAdd) {
	UpdateMemberColorRandom(session, Member.GuildID, Member.User.ID)
	if AutoKick {
		go KickMemberAfterTime(session, Member.GuildID, Member.User.ID)
	}
}

func CheckAdmin(UserID string) (bool) {
	if _, ok := AdminIDs[UserID]; ok {
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

func loadRoles(session *discordgo.Session, GuildID string) {
	GuildRoles, err := session.GuildRoles(GuildID)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		panic("Can't find the server.")
	}

	// Initialise nested map with GuildID as key
	CreatedRoles[GuildID] = map[string]Roles{}
	CreatedRolesID[GuildID] = map[string]Roles{}
	for _, Role := range GuildRoles {
		if _, ok := RoleList[Role.Name]; ok {
			CreatedRoles[GuildID][Role.Name] = Roles{Role.ID, Role.Name}
			CreatedRolesID[GuildID][Role.ID] = Roles{Role.ID, Role.Name}
		}
	}
}

func JoinedNewGuild(session *discordgo.Session, GuildID string) {
	// Initialise nested map with GuildID as key
	CreatedRoles[GuildID] = map[string]Roles{}
	CreatedRolesID[GuildID] = map[string]Roles{}
	fmt.Printf("Joined a new server: %s\n", GuildID)
	CreateAllRoles(session, GuildID)
}

func AddAllMembersColor(session *discordgo.Session, GuildID string) {
	fmt.Printf("Updating all member with new a color.\n")
	session.RequestGuildMembers(GuildID, "", 0)
}

func RemoveAllColors(session *discordgo.Session, GuildID string) {
	GuildRoles, err := session.GuildRoles(GuildID)
	if err != nil {
		panic("Can't find the server.\n")
	}

	for _, Role := range GuildRoles {
		if _, ok := RoleList[Role.Name]; ok {
			session.GuildRoleDelete(GuildID, Role.ID)
		}
	}
}

func UpdateMemberColor(s *discordgo.Session, GuildID, MemberID, RoleName string) {
	s.GuildMemberRoleAdd(GuildID, MemberID, CreatedRoles[GuildID][RoleName].ID)
}

func UpdateMemberColorRandom(s *discordgo.Session, GuildID, MemberID string) {
	key := rand.Intn(len(RoleList))
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
	CreatedRolesID[GuildID][Role.ID] = Roles{Role.ID, Role.Name}
}

func CreateNewRoles(session *discordgo.Session, GuildID string) {
	for Name, Color := range RoleNewList {
		CreateColorRole(session, GuildID, Name, Color)
	}
}

func CreateAllRoles(session *discordgo.Session, GuildID string) {
	for Name, Color := range RoleList {
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
		if _, ok := CreatedRolesID[GuildID][RoleID]; ok {
			session.GuildMemberRoleRemove(GuildID, MemberID, RoleID)
		}
	}
	return false
}

func PreviewRole(session *discordgo.Session, RoleName string) discordgo.MessageEmbed {
	CreateImageWithColor(RoleList[RoleName], RoleName)
	Embed := CreateImageEmbed(session, RoleName)

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
		if _, ok := RoleList[CreatedRolesID[GuildID][RoleID].Name]; ok {
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

	session.ChannelMessageSend(PrivateChannel.ID, "You got kicked from the server. Please read the welcome channel.\n"+InviteLink)
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
