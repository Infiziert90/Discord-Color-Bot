# Discord Color Bot
Adds colored roles to your discord server.

## Work in Progress

## Requirements

To run this bot, you will need:

- Go
- discordgo
- Your own set of Discord credentials to use with the bot (see https://discordapp.com/developers/docs/intro)


## Start

- Open main.go
- Change the settings


## Commands
All commands start with `<<`. Enter parameters after a space (see below for examples).

| Command | Description | Parameter(s) |
| -------- | ----------- | ------------------ |
| Help | Prints a list of all commands |  |
| PrintColors | Prints a list of all colors | |
| NewColor | Assign a random color to the current user |  |
| NewColor | Assign the specified color to the current user | ColorName |
| PreviewColor | Assign the specified color to the bot | ColorName |

### Admin-only commands

| Command | Description | Parameter(s) |
| -------- | ----------- | ------------------ |
| NewServer | Generate all color roles on this server |  |
| AddColorAllMember | Assigns a random color role to all current members | |
| RemoveAllColors | Removes all color roles from the server |  |


Example:
`<<NewColor ColorName`

## Auto Kick
When `var AutoKick` is `true`:

All new members will be kicked after 30 minutes if they do not have at
least one additional role (i. e. a role that is not just a color).
This is a fix for discord’s autokick because the bot gives each user a
role, meaning they won’t be kicked automatically.

## Help?

Add me on discord and message me with your problem:
Infi#8527
