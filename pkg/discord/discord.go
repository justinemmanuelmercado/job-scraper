package discord

import (
	"fmt"
	"log"
	"os"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/bwmarrin/discordgo"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
)

func InitDiscordClient() (*discordgo.Session, error) {
	token := os.Getenv("JOBBYMCJOBFACE_DISCORD_TOKEN")
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("unable to connect discord bot %w", err)
	}

	return bot, nil
}

func convertToSiteUrl(id string) string {
	return fmt.Sprintf("https://workfindy.com/%s", id)
}

func sendEmbed(bot *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) {
	_, err := bot.ChannelMessageSendEmbed(channelID, embed)

	if err != nil {
		log.Printf("Failed to send embed: %v", err) // Log the error or handle it as you prefer
	}
}

func SendList(bot *discordgo.Session, notices []models.Notice) {
	truncateLength := 500
	for _, notice := range notices {
		converter := md.NewConverter("", true, nil)
		notice.Body, _ = converter.ConvertString(notice.Body)

		if len(notice.Body) > truncateLength {
			notice.Body = notice.Body[:truncateLength] + "..."
		}

		// Printf
		notice.Body = fmt.Sprintf("%s\n[View on site](%s)", notice.Body, convertToSiteUrl(notice.ID))

		embed := &discordgo.MessageEmbed{
			Title:       notice.Title,
			URL:         notice.URL,
			Description: notice.Body,
		}

		sendEmbed(bot, os.Getenv("JOBBYMCJOBFACE_HIRING_CHANNEL_ID"), embed)
	}
}

func SendSuccessNotificatioin(bot *discordgo.Session, scrapedCount, newNoticeCount int, runtime time.Duration) error {
	message := fmt.Sprintf("Succesfully run script at: %s\nNotices matched: %d\nNew notices: %d\nRuntime: %v\nSite URL: https://workfindy.com/",
		time.Now().Format("January 2, 2006 15:04:05"), scrapedCount, newNoticeCount, runtime)

	_, err := bot.ChannelMessageSend(os.Getenv("JOBBYMCJOBFACE_HIRING_CHANNEL_ID"), message)
	return err
}
