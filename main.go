package main

import (
	"flag"
	"fmt"
	"go_discord_yelp_food_search/lib"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var botToken string
var guildId string
var RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")

var bot *discordgo.Session

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "yelp-food-search",
		Description: "search food using yelp api",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "location",
				Description: "location",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "categories",
				Description: "categories",
				Required:    true,
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"yelp-food-search": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		inputMap := make(map[string]interface{})
		for _, opt := range options {
			inputMap[opt.Name] = opt.Value
		}
		obj := lib.Search(inputMap["location"].(string), "food", inputMap["categories"].(string))
		// fmt.Println(obj)
		var embeds []*discordgo.MessageEmbed
		for _, d := range obj {
			businessesInfo := d.(map[string]interface{})
			display_address_slice := businessesInfo["location"].(map[string]interface{})["display_address"]
			display_address := ""
			for _, elem := range display_address_slice.([]interface{}) {
				display_address += fmt.Sprint(elem)
			}
			embeds = append(embeds, &discordgo.MessageEmbed{
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Name",
						Value: businessesInfo["name"].(string),
					},
					{
						Name:  "Address",
						Value: display_address,
					},
					{
						Name:  "Rating",
						Value: strconv.FormatFloat(businessesInfo["rating"].(float64), 'f', -1, 64),
					},
					{
						Name:  "Review Count",
						Value: strconv.FormatFloat(businessesInfo["review_count"].(float64), 'f', -1, 64),
					},
					{
						Name:  "Yelp url",
						Value: businessesInfo["url"].(string),
					},
				},
				Image: &discordgo.MessageEmbedImage{
					URL: businessesInfo["image_url"].(string),
				},
			})
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
			},
		})
	},
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	botToken = os.Getenv("DISCORD_BOT_TOKEN")
	guildId = os.Getenv("DISCORD_GUILD_ID")

	bot, err = discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}
	bot.AddHandler(messageCreate)

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	err := bot.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	log.Println("Adding commands...")

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := bot.ApplicationCommandCreate(bot.State.User.ID, guildId, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.

	if *RemoveCommands {
		log.Println("Removing commands...")

		for _, v := range registeredCommands {
			err := bot.ApplicationCommandDelete(bot.State.User.ID, guildId, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	bot.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("test")
}
