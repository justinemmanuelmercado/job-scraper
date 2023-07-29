package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
)

func InitDiscordClient() (*discordgo.Session, error) {
	token := os.Getenv("JOBBYMCJOBFACE_DISCORD_TOKEN")
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("unable to connect discord bot %w", err)
	}

	return bot, nil
}

func SendNotification(message string) error {
	bot, err := InitDiscordClient()
	if err != nil {
		return fmt.Errorf("unable to connect discord bot %w", err)
	}

	_, err = bot.ChannelMessageSend(os.Getenv("JOBBYMCJOBFACE_HIRING_CHANNEL_ID"), message)
	return err
}
