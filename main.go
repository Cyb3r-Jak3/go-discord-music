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
				Name:    lavalinkNodeFlagName,
				Aliases: []string{"l"},
				Usage:   "Lavalink node configuration in the format 'name|address|password'. Note that the address should include the protocol (http or https).",
				Sources: cli.EnvVars("LAVALINK_NODE"),
			},
			&cli.StringSliceFlag{
				Name:    "lavalink_nodes",
				Aliases: []string{"L"},
				Usage:   "Lavalink node configurations in the format 'name|address|password'. Note that the address should include the protocol (http or https). This flag can be used multiple times to specify multiple nodes.",
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
	logger.SetFormatter(&logrus.JSONFormatter{})
	err := SetLogLevel(logger, c.String("log_level"))
	if err != nil {
		return fmt.Errorf("error setting log level: %w", err)
	}
	logger.Infof("Starting go-discord-music bot version %s", version.Version)
	var botOptions []bot.Option
	nodeInfo := c.String(lavalinkNodeFlagName)
	if nodeInfo != "" {
		nodeConfig, nodeConfigErr := LavaLinkNodeString(nodeInfo)
		if nodeConfigErr != nil {
			return fmt.Errorf("error parsing lavalink node configuration: %w", nodeConfigErr)
		}
		logger.Infof("Using custom Lavalink node: %s", nodeConfig.Name)
		botOptions = append(botOptions, bot.WithLavaLinkNode(nodeConfig))
	} else if len(c.StringSlice(lavalinkNodesFlagName)) > 0 {
		for _, node := range c.StringSlice(lavalinkNodesFlagName) {
			nodeConfig, nodeConfigErr := LavaLinkNodeString(node)
			if nodeConfigErr != nil {
				return fmt.Errorf("error parsing lavalink node configuration '%s': %w", node, nodeConfigErr)
			}
			logger.Infof("Using additional Lavalink node: %s", nodeConfig.Name)
			botOptions = append(botOptions, bot.WithLavaLinkNode(nodeConfig))
		}
	}
	b, err := bot.NewBot(c.String("discord_token"), logger, botOptions...)
	if err != nil {
		return fmt.Errorf("error creating bot: %w", err)
	}
	logger.Info("Starting bot...")
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

func LavaLinkNodeString(info string) (disgolink.NodeConfig, error) {
	parts := strings.Split(info, "|")
	if len(parts) != 3 {
		return disgolink.NodeConfig{}, fmt.Errorf("invalid lavalink node format, expected 'name|address|password'. If there is no password, use 'name|address|'")
	}
	secure := parts[1][:5] == "https"

	address := strings.TrimPrefix(parts[1], "http://")
	address = strings.TrimPrefix(address, "https://")

	return disgolink.NodeConfig{
		Name:     parts[0],
		Address:  address,
		Password: parts[2],
		Secure:   secure,
	}, nil
}
