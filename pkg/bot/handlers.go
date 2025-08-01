package bot

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

var bassBoost = &lavalink.Equalizer{
	0:  0.2,
	1:  0.15,
	2:  0.1,
	3:  0.05,
	4:  0.0,
	5:  -0.05,
	6:  -0.1,
	7:  -0.1,
	8:  -0.1,
	9:  -0.1,
	10: -0.1,
	11: -0.1,
	12: -0.1,
	13: -0.1,
	14: -0.1,
}

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)
)

func (b *Bot) shuffle(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	queue := b.Queues.Get(*event.GuildID())
	if queue == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	queue.Shuffle()
	return event.CreateMessage(discord.MessageCreate{
		Content: "Queue shuffled",
	})
}

func (b *Bot) volume(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	volume := data.Int("volume")
	if err := player.Update(context.TODO(), lavalink.WithVolume(volume)); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while setting volume: `%s`", err),
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Volume set to `%d`", volume),
	})
}

func (b *Bot) seek(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	position := data.Int("position")
	unit, ok := data.OptInt("unit")
	if !ok {
		unit = 1
	}
	finalPosition := lavalink.Duration(position * unit)
	if err := player.Update(context.TODO(), lavalink.WithPosition(finalPosition)); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while seeking: `%s`", err),
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Seeked to `%s`", formatPosition(finalPosition)),
	})
}

func (b *Bot) bassBoost(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	enabled := data.Bool("enabled")
	filters := player.Filters()
	if enabled {
		filters.Equalizer = bassBoost
	} else {
		filters.Equalizer = nil
	}

	if err := player.Update(context.TODO(), lavalink.WithFilters(filters)); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while setting bass boost: `%s`", err),
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Bass boost set to `%t`", enabled),
	})
}

func (b *Bot) skip(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	queue := b.Queues.Get(*event.GuildID())
	if player == nil || queue == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	amount, ok := data.OptInt("amount")
	if !ok {
		amount = 1
	}

	track, ok := queue.Skip(amount)
	currentTrack := player.Track()
	if currentTrack != nil {
		playerUpdateErr := player.Update(context.TODO(), lavalink.WithNullTrack())
		if playerUpdateErr != nil {
			return event.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Error while skipping current track: `%s`", playerUpdateErr),
			})
		}
		if !ok {
			b.idleTimes[*event.GuildID()] = time.Now()
			return event.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Skipped `%d` track(s), but no next track available, current track was: [`%s`](<%s>)", amount, currentTrack.Info.Title, *currentTrack.Info.URI),
			})
		}
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Skipped `%d` track(s), current track was: [`%s`](<%s>)", amount, currentTrack.Info.Title, *currentTrack.Info.URI),
		})
	}
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No tracks in queue",
		})
	}

	if err := player.Update(context.TODO(), lavalink.WithTrack(track)); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while skipping track: `%s`", err),
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: "Skipped track",
	})
}

func (b *Bot) queueType(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	queue := b.Queues.Get(*event.GuildID())
	if queue == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	queue.Type = QueueType(data.String("type"))
	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Queue type set to `%s`", queue.Type),
	})
}

func (b *Bot) clearQueue(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	queue := b.Queues.Get(*event.GuildID())
	if queue == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	queue.Clear()
	return event.CreateMessage(discord.MessageCreate{
		Content: "Queue cleared",
	})
}

func (b *Bot) queue(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	queue := b.Queues.Get(*event.GuildID())
	if queue == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No queue found",
		})
	}

	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	var tracks string
	currentTrack := player.Track()
	if currentTrack != nil {
		tracks += fmt.Sprintf("Current track: [`%s`](<%s>)\n", currentTrack.Info.Title, *currentTrack.Info.URI)
	} else {
		tracks += "No current track\n"
	}

	if len(queue.Tracks) == 0 {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No tracks in queue",
		})
	}

	for i, track := range queue.Tracks {
		tracks += fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, track.Info.Title, *track.Info.URI)
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Queue `%s`:\n%s", queue.Type, tracks),
	})
}

func (b *Bot) players(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	var description string
	b.Lavalink.ForPlayers(func(player disgolink.Player) {
		description += fmt.Sprintf("GuildID: `%s`\n", player.GuildID())
	})

	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Players:\n%s", description),
	})
}

func (b *Bot) pause(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	if err := player.Update(context.TODO(), lavalink.WithPaused(!player.Paused())); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while pausing: `%s`", err),
		})
	}

	status := "playing"
	if player.Paused() {
		status = "paused"
	}
	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Player is now %s", status),
	})
}

func (b *Bot) stop(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	if err := player.Update(context.TODO(), lavalink.WithNullTrack()); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while stopping: `%s`", err),
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: "Player stopped",
	})
}

func (b *Bot) connect(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "Player already connected",
		})
	}
	voiceState, ok := b.Client.Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Content: "You need to be in a voice channel to use this command",
		})
	}
	err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), voiceState.ChannelID, false, false)
	if err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while connecting: `%s`", err),
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: "Player connected",
	})
}

func (b *Bot) disconnect(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	if err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), nil, false, false); err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error while disconnecting: `%s`", err),
		})
	}
	delete(b.idleTimes, *event.GuildID())

	return event.CreateMessage(discord.MessageCreate{
		Content: "Player disconnected",
	})
}

func (b *Bot) nowPlaying(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	player := b.Lavalink.ExistingPlayer(*event.GuildID())
	if player == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No player found",
		})
	}

	track := player.Track()
	if track == nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: "No track found",
		})
	}

	return event.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Now playing: [`%s`](<%s>)\n\n %s / %s", track.Info.Title, *track.Info.URI, formatPosition(player.Position()), formatPosition(track.Info.Length)),
	})
}

func formatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}

func (b *Bot) play(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error {
	identifier := data.String("identifier")
	if source, ok := data.OptString("source"); ok {
		identifier = lavalink.SearchType(source).Apply(identifier)
	} else if !urlPattern.MatchString(identifier) && !searchPattern.MatchString(identifier) {
		identifier = lavalink.SearchTypeYouTube.Apply(identifier)
	}

	voiceState, ok := b.Client.Caches().VoiceState(*event.GuildID(), event.User().ID)
	if !ok {
		return event.CreateMessage(discord.MessageCreate{
			Content: "You need to be in a voice channel to use this command",
		})
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track
	b.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			_, _ = b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: common.Ptr(fmt.Sprintf("Loaded track: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
			})
			toPlay = &track
		},
		func(playlist lavalink.Playlist) {
			_, _ = b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: common.Ptr(fmt.Sprintf("Loaded playlist: `%s` with `%d` tracks", playlist.Info.Name, len(playlist.Tracks))),
			})
			toPlay = &playlist.Tracks[0]
		},
		func(tracks []lavalink.Track) {
			_, _ = b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: common.Ptr(fmt.Sprintf("Loaded search result: [`%s`](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI)),
			})
			toPlay = &tracks[0]
		},
		func() {
			_, _ = b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: common.Ptr(fmt.Sprintf("Nothing found for: `%s`", identifier)),
			})
		},
		func(err error) {
			_, _ = b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: common.Ptr(fmt.Sprintf("Error while looking up query: `%s`", err)),
			})
		},
	))
	if toPlay == nil {
		return nil
	}

	if err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), voiceState.ChannelID, false, false); err != nil {
		return err
	}
	b.logger.Infof("Found track: %s", toPlay.Info.Title)

	playErr := b.Lavalink.Player(*event.GuildID()).Update(context.TODO(), lavalink.WithTrack(*toPlay))
	if playErr != nil {
		_, _ = b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: common.Ptr(fmt.Sprintf("Error while playing track: `%s`", playErr)),
		})
		return playErr
	}

	return playErr
}

func (b *Bot) debug(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	selfInfo, err := b.Client.Rest().GetCurrentApplication()
	if err != nil {
		return event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Error fetching application info: `%s`", err),
		})
	}

	if event.User().ID != selfInfo.Owner.ID {
		return event.CreateMessage(discord.MessageCreate{
			Content: "You are not allowed to use this command",
		})
	}

	var gatewayPing string
	if event.Client().HasGateway() {
		gatewayPing = event.Client().Gateway().Latency().String()
	}
	eb := discord.NewEmbedBuilder().
		SetTitle("Debug Info").
		AddField("Gateway", gatewayPing, false)

	cookieString := ""
	nodeString := ""
	b.Lavalink.ForNodes(func(node disgolink.Node) {
		nodeConfig := node.Config()
		nodeHost, parseErr := url.Parse(nodeConfig.Address)
		if parseErr != nil {
			b.logger.Errorf("error parsing node address: %v", parseErr)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		nodeVersion, nodeErr := node.Version(ctx)
		cancel()
		if nodeErr != nil {
			b.logger.Errorf("error getting node version: %v", nodeErr)
			nodeVersion = "error"
		}
		nodeString += fmt.Sprintf("Node: `%s` Address: `%s` Secure: `%t` Version: %s\n", nodeConfig.Name, nodeConfig.Address, nodeConfig.Secure, nodeVersion)
		cookies := b.HTTPClient.Jar.Cookies(nodeHost)
		if len(cookies) == 0 {
			cookieString += fmt.Sprintf("Node: `%s` (%s) has no cookies\n", nodeConfig.Name, nodeHost)
			return
		}
		for _, cookie := range cookies {
			cookieString += fmt.Sprintf("Node: `%s` Cookie: `%s` Value: `%s`\n", nodeConfig.Name, cookie.Name, cookie.Value)
		}
	})
	timerString := ""
	for guild, timer := range b.idleTimes {
		if time.Since(timer) > b.IdleTimeout {
			disconnectErr := b.Client.UpdateVoiceState(context.TODO(), guild, nil, false, false)
			if disconnectErr != nil {
				b.logger.Errorf("error updating voice state for guild %s: %v", guild, disconnectErr)
			}
			delete(b.idleTimes, guild)
			b.logger.Infof("Guild `%s` has been idle for more than %s, disconnected\n", guild, b.IdleTimeout)
		} else {
			b.logger.Infof("idle timeout for guild %s, %d", guild, timer.Unix())
			timerString += fmt.Sprintf("Guild `%s`: idle time remaining: %s\n", guild, b.IdleTimeout-time.Since(timer).Round(time.Second))
		}
	}
	if timerString == "" {
		timerString = "No guilds are idle"
	}
	eb.AddField("Idle Times", timerString, false)
	eb.AddField("Nodes", nodeString, false)
	eb.AddField("HTTP Client Cookies", cookieString, false)
	eb.Timestamp = common.Ptr(time.Now())
	return event.Respond(discord.InteractionResponseTypeCreateMessage, discord.NewMessageCreateBuilder().
		SetEmbeds(eb.Build()).
		Build(),
	)
}

func (b *Bot) source(event *events.ApplicationCommandInteractionCreate, _ discord.SlashCommandInteractionData) error {
	return event.CreateMessage(discord.MessageCreate{
		Content: "Source for the bot: [Cyb3r-Jak3/go-discord-music](https://github.com/Cyb3r-Jak3/go-discord-music)\n",
	})
}
