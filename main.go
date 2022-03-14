package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/sadlil/go-trigger"
	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

var discordConfig, ircConfig, mapToDiscord, mapToIRC map[string]interface{}

type msgObject struct {
	Target  string
	Channel string
	Message string
	User    string
	_User   string
	__User  string
}

func loadConfig() {
	viper.SetConfigName("config.json")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s ", err))
	} else {
		fmt.Println("Config file successfully read")
	}
	ircConfig = viper.Get("irc").(map[string]interface{})
	discordConfig = viper.Get("discord").(map[string]interface{})
	mapToDiscord = viper.GetStringMap("channelMapping")
	mapToIRC = reverseChannelMap(mapToDiscord)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content != "" {
		trigger.Fire("core::message", &msgObject{
			Target:  "irc",
			Channel: m.ChannelID,
			Message: m.Content,
			User:    m.Member.Nick,
			_User:   m.Author.ID,
			__User:  m.Author.Username,
		})
	}

	if m.Attachments != nil {
		for _, attachment := range m.Attachments {
			if attachment.Size != 0 {
				newURL := getShortLink(attachment.URL)
				trigger.Fire("core::message", &msgObject{
					Target:  "irc",
					Channel: m.ChannelID,
					Message: newURL,
					User:    m.Member.Nick,
					_User:   m.Author.ID,
					__User:  m.Author.Username,
				})
			}
		}
	}
}

func main() {

	confCheck := flag.Bool("confcheck", false, "Validates config file, exits after completion")

	flag.Parse()

	if *confCheck {
		loadConfig()
		os.Exit(3)
	}

	loadConfig()
	ircClient := irc.IRC(ircConfig["nick"].(string), ircConfig["nick"].(string))
	ircClient.Connect(ircConfig["server"].(string) + ":" + ircConfig["port"].(string))

	ircClient.AddCallback("001", func(e *irc.Event) {
		ircClient.Join("#test")

		for _, channel := range mapToIRC {
			ircClient.Join(channel.(string))
		}
	})

	ircClient.AddCallback("PRIVMSG", func(e *irc.Event) {
		trigger.Fire("core::message", &msgObject{
			Target:  "discord",
			Channel: e.Arguments[0],
			Message: e.Message(),
			User:    e.Nick,
			_User:   e.Nick,
			__User:  e.Nick,
		})
	})

	discordClient, err := discordgo.New("Bot " + viper.Get("discord.authToken").(string))
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	discordClient.AddHandler(messageCreate)

	discordClient.Identify.Intents = discordgo.IntentsGuildMessages
	err = discordClient.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	trigger.On("core::message", func(data interface{}) {
		fmt.Println("Got message: ", data)
		msg := data.(*msgObject)

		if !channelIsInMap(msg.Channel) {
			fmt.Println("Channel not in map")
			return
		}

		fmt.Println("Channel is in map")

		var target string = ""

		if msg.Target == "discord" {
			target = mapToDiscord[msg.Channel].(string)
			discordClient.ChannelMessageSend(target, formatter(msg))
		} else if msg.Target == "irc" {
			target = mapToIRC[msg.Channel].(string)
			ircClient.Privmsg(target, formatter(msg))
		}
	})

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	ircClient.Loop()
	discordClient.Close()

}
