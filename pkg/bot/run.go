package bot

import (
	"context"
	"go-discord-music/pkg/version"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
)

func (b *Bot) Run() {
	b.logger.Debugf("starting bot with version: %s", version.Version)
	b.registerCommands()

	b.Lavalink = disgolink.New(b.Client.ApplicationID(),
		disgolink.WithListenerFunc(b.onPlayerPause),
		disgolink.WithListenerFunc(b.onPlayerResume),
		disgolink.WithListenerFunc(b.onTrackStart),
		disgolink.WithListenerFunc(b.onTrackEnd),
		disgolink.WithListenerFunc(b.onTrackException),
		disgolink.WithListenerFunc(b.onTrackStuck),
		disgolink.WithListenerFunc(b.onWebSocketClosed),
		disgolink.WithListenerFunc(b.onUnknownEvent),
		disgolink.WithLogger(slog.New(NewLogrusAdapter(b.logger))),
		disgolink.WithHTTPClient(b.HTTPClient),
	)
	nodeCount := 0
	b.Lavalink.ForNodes(func(_ disgolink.Node) { nodeCount++ })
	if nodeCount == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, addNodeErr := b.Lavalink.AddNode(ctx, defaultLavalinkNode)
		if addNodeErr != nil {
			b.logger.Fatalf("error adding default lavalink node: %v", addNodeErr)
		}
	}
	b.Handlers = map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error{
		"play":        b.play,
		"pause":       b.pause,
		"now-playing": b.nowPlaying,
		"stop":        b.stop,
		"players":     b.players,
		"queue":       b.queue,
		"clear-queue": b.clearQueue,
		"queue-type":  b.queueType,
		"shuffle":     b.shuffle,
		"seek":        b.seek,
		"volume":      b.volume,
		"skip":        b.skip,
		"bass-boost":  b.bassBoost,
		"disconnect":  b.disconnect,
		"connect":     b.connect,
		"debug":       b.debug,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.Client.OpenGateway(ctx); err != nil {
		b.logger.Fatalf("error opening discord gateway: %v", err)
	}
}
