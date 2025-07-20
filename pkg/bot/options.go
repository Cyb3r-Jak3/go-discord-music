package bot

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/sirupsen/logrus"
)

var (
	defaultLavalinkNode = disgolink.NodeConfig{
		Name:     "cyberjake-default",
		Address:  "lavalink-4-lb.cyberjake.xyz",
		Password: "FreeToUse",
		Secure:   true,
	}
	lavalinkSingleton = disgolink.NodeConfig{
		Name:     "cyberjake-default",
		Address:  "lavalink-4.cyberjake.xyz",
		Password: "FreeToUse",
		Secure:   true,
	}
)

// Option is a functional option for configuring the API client.
type Option func(*Bot) error

func WithLogger(logger *logrus.Logger) Option {
	return func(b *Bot) error {
		if logger == nil {
			return fmt.Errorf("logger is required when using WithLogger option")
		}
		b.logger = logger
		return nil
	}
}

func WithLavaLinkNode(node disgolink.NodeConfig) Option {
	return func(b *Bot) error {
		if b.Lavalink == nil {
			return fmt.Errorf("lavalink client is required when using WithLavaLinkNode option")
		}
		_, err := b.Lavalink.AddNode(context.Background(), node)
		return err
	}
}

func WithLavaLinkSingleton() Option {
	return func(b *Bot) error {
		if b.Lavalink == nil {
			return fmt.Errorf("lavalink client is required when using WithLavaLinkSingleton option")
		}
		_, err := b.Lavalink.AddNode(context.Background(), lavalinkSingleton)
		return err
	}
}
