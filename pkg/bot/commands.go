package bot

import (
	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

var commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "play",
		Description: "Plays a song",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "identifier",
				Description: "The song link or search query",
				Required:    true,
			},
			discord.ApplicationCommandOptionString{
				Name:        "source",
				Description: "The source to search on",
				Required:    false,
				Choices: []discord.ApplicationCommandOptionChoiceString{
					{
						Name:  "YouTube",
						Value: string(lavalink.SearchTypeYouTube),
					},
					{
						Name:  "YouTube Music",
						Value: string(lavalink.SearchTypeYouTubeMusic),
					},
					{
						Name:  "SoundCloud",
						Value: string(lavalink.SearchTypeSoundCloud),
					},
					{
						Name:  "Deezer",
						Value: "dzsearch",
					},
					{
						Name:  "Deezer ISRC",
						Value: "dzisrc",
					},
					{
						Name:  "Spotify",
						Value: "spsearch",
					},
					{
						Name:  "AppleMusic",
						Value: "amsearch",
					},
				},
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "pause",
		Description: "Pauses the current song",
	},
	discord.SlashCommandCreate{
		Name:        "now-playing",
		Description: "Shows the current playing song",
	},
	discord.SlashCommandCreate{
		Name:        "stop",
		Description: "Stops the current song and stops the player",
	},
	discord.SlashCommandCreate{
		Name:        "disconnect",
		Description: "Disconnects the player",
	},
	discord.SlashCommandCreate{
		Name:        "bass-boost",
		Description: "Enables or disables bass boost",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "enabled",
				Description: "Whether bass boost should be enabled or disabled",
				Required:    true,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "players",
		Description: "Shows all active players",
	},
	discord.SlashCommandCreate{
		Name:        "skip",
		Description: "Skips the current song",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "amount",
				Description: "The amount of songs to skip",
				Required:    false,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "volume",
		Description: "Sets the volume of the player",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "volume",
				Description: "The volume to set",
				Required:    true,
				MaxValue:    common.Ptr(1000),
				MinValue:    common.Ptr(0),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "seek",
		Description: "Seeks to a specific position in the current song",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "position",
				Description: "The position to seek to",
				Required:    true,
			},
			discord.ApplicationCommandOptionInt{
				Name:        "unit",
				Description: "The unit of the position",
				Required:    false,
				Choices: []discord.ApplicationCommandOptionChoiceInt{
					{
						Name:  "Milliseconds",
						Value: int(lavalink.Millisecond),
					},
					{
						Name:  "Seconds",
						Value: int(lavalink.Second),
					},
					{
						Name:  "Minutes",
						Value: int(lavalink.Minute),
					},
					{
						Name:  "Hours",
						Value: int(lavalink.Hour),
					},
				},
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "shuffle",
		Description: "Shuffles the current queue",
	},
	discord.SlashCommandCreate{
		Name:        "connect",
		Description: "Forces the bot to connect to a voice channel",
	},
	discord.SlashCommandCreate{
		Name:        "debug",
		Description: "Debug command to get information about the bot",
	},
	discord.SlashCommandCreate{
		Name:        "source",
		Description: "GitHub link to the source code of the bot",
	},
}

func (b *Bot) registerCommands() {
	if err := handler.SyncCommands(b.Client, commands, []snowflake.ID{}); err != nil {
		b.logger.Fatalf("error while registering commands: %v", err)
	}
}
