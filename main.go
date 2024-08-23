package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

func main() {
	fmt.Print("Please enter an xkcd comic number: ")
	var number string
	fmt.Scan(&number)

	content := getHTML("https://xkcd.com/" + number)
	re := regexp.MustCompile(`https://imgs.xkcd.com/comics/.*\.png`)
	link := re.FindString(content)
	fmt.Println(link)
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
