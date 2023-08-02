package discord

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func InitDiscordClient() (*discordgo.Session, error) {
	token := os.Getenv("JOBBYMCJOBFACE_DISCORD_TOKEN")
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("unable to connect discord bot %w", err)
	}

	return bot, nil
}

func SendNotification(noticeCount, newNoticeCount int, runtime time.Duration) error {
	bot, err := InitDiscordClient()
	if err != nil {
		return fmt.Errorf("unable to connect discord bot %w", err)
	}

	message := fmt.Sprintf("Succesfully run script at: %s\nNotices matched: %d\nNew notices: %d\nRuntime: %v",
		time.Now().Format("January 2, 2006 15:04:05"), noticeCount, newNoticeCount, runtime)

	_, err = bot.ChannelMessageSend(os.Getenv("JOBBYMCJOBFACE_HIRING_CHANNEL_ID"), message)
	return err
}
