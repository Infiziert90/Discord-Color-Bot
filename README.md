# Discord Color Bot
Adds colored roles to your discord server.

## Work in Progress

## Requirements

To run this bot, you will need:

- Go
- discordgo
- Your own set of Discord credentials to use with the bot (see https://discordapp.com/developers/docs/intro)


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

## Help?

Add me on discord and message me with your problem:
Infi#8527
