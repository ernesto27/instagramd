package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage ./instagramd URLIMAGE")
		return
	}

	htmlMeta, err := getHTMLMeta(os.Args[1])
	if err != nil {
		log.Fatalln(err.Error())
	}

	success, err := downloadFile(htmlMeta)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(success)

}

func downloadFile(htmlMeta *HTMLMeta) (string, error) {
	var url string
	var filenameExt string
	if htmlMeta.Video != "" {
		url = htmlMeta.Video
		filenameExt = "mp4"
	} else {
		url = htmlMeta.Image
		filenameExt = "jpeg"
	}

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.New("Received non 200 response code")
	}

	filename := getRandomString() + "." + filenameExt
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return "Success download filename: " + filename, nil
}

func getHTMLMeta(url string) (*HTMLMeta, error) {
	req, err := http.NewRequest("GET", url, nil)
	htmlMeta := HTMLMeta{}
	if err != nil {
		return &htmlMeta, err
	}
	req.Header.Set("User-Agent", "Instagram 10.3.2 (iPhone7,2; iPhone OS 9_3_3; en_US; en-US; scale=2.00; 750x1334) AppleWebKit/420+")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &htmlMeta, err
	}

	if resp.StatusCode != 200 {
		return &htmlMeta, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	meta := extract(resp.Body)
	return meta, nil
}

type HTMLMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
	Video       string `json:"video"`
}

func extract(resp io.Reader) *HTMLMeta {
	z := html.NewTokenizer(resp)

	titleFound := false

	hm := new(HTMLMeta)

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return hm
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == `body` {
				return hm
			}
			if t.Data == "title" {
				titleFound = true
			}
			if t.Data == "meta" {
				desc, ok := extractMetaProperty(t, "description")
				if ok {
					hm.Description = desc
				}

				ogTitle, ok := extractMetaProperty(t, "og:title")
				if ok {
					hm.Title = ogTitle
				}

				ogDesc, ok := extractMetaProperty(t, "og:description")
				if ok {
					hm.Description = ogDesc
				}

				ogImage, ok := extractMetaProperty(t, "og:image")
				if ok {
					hm.Image = ogImage
				}

				ogSiteName, ok := extractMetaProperty(t, "og:site_name")
				if ok {
					hm.SiteName = ogSiteName
				}

				ogVideo, ok := extractMetaProperty(t, "og:video")
				if ok {
					hm.Video = ogVideo
				}
			}
		case html.TextToken:
			if titleFound {
				t := z.Token()
				hm.Title = t.Data
				titleFound = false
			}
		}
	}
}

func extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "property" && attr.Val == prop {
			ok = true
		}

		if attr.Key == "content" {
			content = attr.Val
		}
	}
	return
}

func getRandomString() string {
	rand.Seed(time.Now().Unix())
	var output strings.Builder
	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	length := 20
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}
