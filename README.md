# Discord Color Bot
Adds colored roles to your discord server.

## Work in Progress
#### Admin commands are currently not implemented

## Requirements

To run this bot, you will need:

- Rust
- Your own set of Discord credentials to use with the bot (see https://discordapp.com/developers/docs/intro)


## Start

- Edit the config.yaml
- Compile the rust project
- Enjoy


## Commands
All commands are available as slash-command or as chat commands with `<<`.  
Enter parameters after a space. (see below for an example)

| Command | Description                   | Parameter(s) |
|---------|-------------------------------|--------------|
| help    | Prints a list of all commands |              |
| color   | Assign a random color         |              |
| color   | Assign the specified color    | Color Name   |
| preview | Send a preview image          | Color Name   |

### NOT REIMPLEMENTED YET

| Command             | Description                                        | Parameter(s) |
|---------------------|----------------------------------------------------|--------------|
| NewServer           | Generate all color roles on this server            |              |
| AddColorToAllMember | Assigns a random color role to all current members |              |
| RemoveAllColors     | Removes all color roles from the server            |              |
| ReloadColors        | Add all new colors to your server                  |              |


Example:
`<<color Fuchsia`

## Usage
### Preview
![preview.gif](images/preview.gif)

### Assignment
![preview.gif](images/color.gif)

## Adding Colors
In your config file, you can find a list named “Colors”, there you can add new colors.  
You have to restart the bot afterward.

## Auto Kick
If you want it, just add your ServerID into `AutoKickOnServer`

All new members will be kicked after 30 minutes, if they do not have at
least one additional role (i.e. a role that isn't a color).
This is a fix for discord’s auto kick, as the bot gives each user a
role, so they won’t be kicked automatically.

## Help?

Add me on discord and message me with your problem:  
@infi
