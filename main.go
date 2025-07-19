package main

import (
	"context"
	"fmt"
	"go-discord-music/pkg/bot"
	"go-discord-music/pkg/version"
	"net/mail"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func buildApp() *cli.Command {
	versionString := version.String()
	app := &cli.Command{
		Name:    "go-discord-music",
		Usage:   "A simple Discord music bot",
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
	logger.SetLevel(logrus.DebugLevel)
	b, err := bot.NewBot(c.String("discord_token"), bot.WithLogger(logger))
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
