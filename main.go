package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

const comicPrefix = "https://xkcd.com/"
const randomLink = "https://c.xkcd.com/random/comic/"

var comicNumberRegex = regexp.MustCompile(`Permanent link to this comic: <a href="https://xkcd.com/(.*?)">`)
var comicLinkRegex = regexp.MustCompile(`Permanent link to this comic: <a href="(.*?)">`)

var (
	BotToken       = os.Getenv("AUTH_TOKEN")
	RemoveCommands = true
)

func init() {
	flag.Parse()
}

var s *discordgo.Session

func init() {
	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "xkcd",
			Description: "Get an xkcd",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Comic ID",
					MinValue:    &integerOptionMinValue,
					MaxValue:    69000,
					Required:    true,
				},
			},
		},
		{
			Name:        "xkcd-latest",
			Description: "Get the latest xkcd",
		},
		{
			Name:        "xkcd-random",
			Description: "Get a random xkcd",
		},
		{
			Name:        "xkcd-standards",
			Description: "14 + 1 standards",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"xkcd": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			id := options[0].IntValue()

			if id > getLatestComicID() {
				respond(s, i, "Woah there! That comic hasn't been written yet. If you read it in an alternate time stream, please contact the appropriate authorities")
			} else {
				link := getComic(id)
				respond(s, i, link)
			}
		},
		"xkcd-latest": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			latest := getLatestComicID()
			link := getComic(latest)
			respond(s, i, link)
		},
		"xkcd-random": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			link := getRandomComic()
			respond(s, i, link)
		},
		"xkcd-standards": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			link := getComic(927)
			respond(s, i, link)
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func getLatestComicID() int64 {
	content := getHTML("https://xkcd.com/")
	matches := comicNumberRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		if id, err := strconv.Atoi(matches[1]); err == nil {
			log.Printf("Latest comic ID: %d", id)
			return int64(id)
		}
	}

	// The latest comic when this bot was created was 2976, so this is a safe and useful default
	log.Println("Could not get latest comic ID, defaulting to 2976")
	return 2976
}

func getComic(number int64) string {
	content := getHTML(comicPrefix + strconv.FormatInt(number, 10))
	return getComicLink(&content)
}

func getRandomComic() string {
	content := getHTML(randomLink)
	return getComicLink(&content)
}

func getHTML(link string) string {
	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func getComicLink(content *string) string {
	matches := comicLinkRegex.FindStringSubmatch(*content)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ""
	}
}
