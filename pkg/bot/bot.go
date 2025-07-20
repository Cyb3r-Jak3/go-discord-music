package bot

import (
	"context"
	"errors"
	"fmt"
	"go-discord-music/pkg/version"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	Client      bot.Client
	Lavalink    disgolink.Client
	Handlers    map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData) error
	Queues      *QueueManager
	HTTPClient  *http.Client
	logger      *logrus.Logger
	VersionInfo string
}

func NewBot(Token string, logger *logrus.Logger, opts ...Option) (*Bot, error) {
	b := &Bot{
		Queues: &QueueManager{
			queues: make(map[snowflake.ID]*Queue),
		},
		logger:      logger,
		VersionInfo: version.String(),
	}

	// Cookie jar is needed as the default Lavalink node is proxied and uses sticky session
	jar, err := cookiejar.New(nil)
	if err != nil {
		b.logger.Fatalf("error creating cookie jar: %v", err)
	}
	httpClient := &http.Client{
		Jar: jar,
	}
	b.HTTPClient = httpClient

	client, err := disgo.New(Token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
		),
		bot.WithEventListenerFunc(b.onApplicationCommand),
		bot.WithEventListenerFunc(b.onVoiceStateUpdate),
		bot.WithEventListenerFunc(b.onVoiceServerUpdate),
	)

	if err != nil {
		return nil, errors.Join(fmt.Errorf("error creating the bot client"), err)
	}

	b.Client = client

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
	err = b.parseOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("options parsing failed: %w", err)
	}

	return b, nil
}

func (b *Bot) onApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()

	handler, ok := b.Handlers[data.CommandName()]
	if !ok {
		b.logger.Warnf("unknown command: %s", data.CommandName())
		return
	}
	if err := handler(event, data); err != nil {
		b.logger.Errorf("error handling command %s: %v", data.CommandName(), err)
	}
}

func (b *Bot) onVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.UserID != b.Client.ApplicationID() {
		return
	}
	b.Lavalink.OnVoiceStateUpdate(context.TODO(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
	if event.VoiceState.ChannelID == nil {
		b.Queues.Delete(event.VoiceState.GuildID)
	}
}

func (b *Bot) onVoiceServerUpdate(event *events.VoiceServerUpdate) {
	b.Lavalink.OnVoiceServerUpdate(context.TODO(), event.GuildID, event.Token, *event.Endpoint)
}

func (b *Bot) Shutdown() {
	b.logger.Infof("shutting down...")
	for _, queue := range b.Queues.queues {
		queue.Clear()
	}
	b.logger.Debugf("queues cleared")
	b.Lavalink.Close()
	b.logger.Debugf("lavalink connection closed")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	b.Client.Close(ctx)
	b.logger.Debugf("Bot shutdown complete")
}

// parseOptions parses the supplied options functions and returns a configured
// *API instance.
func (b *Bot) parseOptions(opts ...Option) error {
	// Range over each options function and apply it to our API type to
	// configure it. Options functions are applied in order, with any
	// conflicting options overriding earlier calls.
	for _, option := range opts {
		err := option(b)
		if err != nil {
			return err
		}
	}

	return nil
}
