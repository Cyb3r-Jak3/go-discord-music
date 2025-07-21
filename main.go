package main

import (
	"context"
	"fmt"
	"net/mail"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"go-discord-music/pkg/bot"
	"go-discord-music/pkg/version"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func buildApp() *cli.Command {
	versionString := version.String()
	app := &cli.Command{
		Name:    "go-discord-music",
		Usage:   "Simple Discord music bot",
		Version: versionString,
		Suggest: true,
		Authors: []any{
			&mail.Address{
				Name:    "Cyber-Jak3",
				Address: "git@cyberjake.xyz",
			},
		},
		Action: Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "discord_token",
				Aliases: []string{"t"},
				Usage:   "Discord bot token",
				Sources: cli.EnvVars("DISCORD_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "lavalink_node",
				Aliases: []string{"l"},
				Usage:   "Lavalink node configuration in the format 'name|address|password'. Note that the address should include the protocol (http or https).",
				Sources: cli.EnvVars("LAVALINK_NODE"),
			},
			&cli.BoolFlag{
				Name:    "lavalink_singleton",
				Aliases: []string{"s"},
				Usage:   "Use a singleton Lavalink node. This will use a predefined node configuration.",
				Sources: cli.EnvVars("LAVALINK_SINGLETON"),
			},
			&cli.StringFlag{
				Name:    "log_level",
				Aliases: []string{"v"},
				Usage:   "Set the log level (debug, info, warn, error, fatal, panic). Default is 'warn'.",
				Value:   "warn",
				Sources: cli.EnvVars("LOG_LEVEL"),
			},
		},
		EnableShellCompletion: true,
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	return app
}

func main() {
	app := buildApp()
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Printf("Error running app: %s\n", err)
		os.Exit(1)
	}
}

func Run(_ context.Context, c *cli.Command) error {
	logger := logrus.New()
	err := SetLogLevel(logger, strings.ToUpper(c.String("log_level")))
	if err != nil {
		return fmt.Errorf("error setting log level: %w", err)
	}
	logger.SetLevel(logrus.DebugLevel)
	var botOptions []bot.Option
	nodeInfo := c.String("lavalink_node")
	if nodeInfo != "" {
		parts := strings.Split(nodeInfo, "|")
		if len(parts) != 3 {
			return fmt.Errorf("invalid lavalink node format, expected 'name|address|password'. If there is no password, use 'name|address|'")
		}
		secure := true //
		if parts[1][:5] == "http:" {
			secure = false
		}
		address := strings.TrimPrefix(parts[1], "http://")
		address = strings.TrimPrefix(address, "https://")
		logger.Debugf("Connecting to Lavalink node: Name:%s, Address: %s, Secure: %t", parts[0], address, secure)

		nodeConfig := disgolink.NodeConfig{
			Name:     parts[0],
			Address:  address,
			Password: parts[2],
			Secure:   secure,
		}

		logger.Infof("Using custom Lavalink node: %s", nodeConfig.Name)
		botOptions = append(botOptions, bot.WithLavaLinkNode(nodeConfig))
	}
	if c.Bool("lavalink_singleton") {
		if nodeInfo != "" {
			logger.Warnf("Lavalink node configuration with singleton node. Ignoring --lavalink-singleton flag")
		} else {
			logger.Info("Using default singleton Lavalink node configuration")
			botOptions = append(botOptions, bot.WithLavaLinkSingleton())
		}
	}
	b, err := bot.NewBot(c.String("discord_token"), logger, botOptions...)
	if err != nil {
		return fmt.Errorf("error creating bot: %w", err)
	}
	logger.Info("Bot initialized")
	b.Run()

	// Wait for interrupt signal to gracefully shut down the bot
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	// Perform any necessary cleanup here
	b.Shutdown()
	return nil
}

func SetLogLevel(logger *logrus.Logger, logLevel string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s, error: %w", logLevel, err)
	}
	logger.SetLevel(level)

	logger.Debugf("Log Level set to %v", logger.Level)
	return nil
}
