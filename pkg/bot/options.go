package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
)

var (
	defaultLavalinkNode = disgolink.NodeConfig{
		Name:     "cyberjake-default",
		Address:  "lavalink-4.cyberjake.xyz",
		Password: "FreeToUse",
		Secure:   true,
	}
)

// Option is a functional option for configuring the API client.
type Option func(*Bot) error

func WithLavaLinkNode(node disgolink.NodeConfig) Option {
	return func(b *Bot) error {
		if b.Lavalink == nil {
			return fmt.Errorf("lavalink client is required when using WithLavaLinkNode option")
		}
		b.logger.Debugf("Using Lavalink node configuration: %#v", node)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := b.Lavalink.AddNode(ctx, node)
		if err != nil {
			return fmt.Errorf("error adding lavalink node %s: %w", node.Name, err)
		}
		return err
	}
}

func WithLavaLinkDefault() Option {
	return func(b *Bot) error {
		if b.Lavalink == nil {
			return fmt.Errorf("lavalink client is required when using WithLavaLinkDefault option")
		}
		b.logger.Debugf("Using default Lavalink node configuration: %s", defaultLavalinkNode.Name)
		_, err := b.Lavalink.AddNode(context.Background(), defaultLavalinkNode)
		return err
	}
}

func WithIdleTimeout(timout time.Duration) Option {
	return func(b *Bot) error {
		b.IdleTimeout = timout
		return nil
	}
}
