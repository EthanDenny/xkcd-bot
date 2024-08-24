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

// TODO: Determine this from the home page
var ComicsCreated float64 = 2976.0

var comicLinkRegex = regexp.MustCompile(`Permanent link to this comic: <a href="(.*?)">`)

var (
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = true
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "xkcd",
			Description: "xkcd comics",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Comic ID",
					MinValue:    &integerOptionMinValue,
					MaxValue:    ComicsCreated,
					Required:    true,
				},
			},
		},
		{
			Name:        "xkcd-random",
			Description: "A random comic",
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
			link := getComic(id)
			respondWithComic(s, i, link)
		},
		"xkcd-random": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			link := getRandomComic()
			respondWithComic(s, i, link)
		},
		"xkcd-standards": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			link := getComic(927)
			respondWithComic(s, i, link)
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

func respondWithComic(s *discordgo.Session, i *discordgo.InteractionCreate, link string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: link,
		},
	})
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
