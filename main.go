package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

const comicPrefix = "https://xkcd.com/"
const randomLink = "https://c.xkcd.com/random/comic/"

var comicNameRegex = regexp.MustCompile(`<div id="ctitle">(.*?)</div>`)
var imageLinkRegex = regexp.MustCompile(`https://imgs.xkcd.com/comics/.*\.(png|jpg)`)

func main() {
	name, link := getRandomComic()
	fmt.Println(name)
	fmt.Println(link)
}

func getComicName(content *string) string {
	matches := comicNameRegex.FindStringSubmatch(*content)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ""
	}
}

func getImageLink(content *string) string {
	return imageLinkRegex.FindString(*content)
}

func getComic(number int) (string, string) {
	content := getHTML(comicPrefix + strconv.Itoa(number))
	return getComicDetails(&content)
}

func getRandomComic() (string, string) {
	content := getHTML(randomLink)
	return getComicDetails(&content)
}

func getComicDetails(content *string) (string, string) {
	name := getComicName(content)
	link := getImageLink(content)
	return name, link
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
