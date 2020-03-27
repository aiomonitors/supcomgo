package supcomgo

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	discord "supcomgo/embeds"

	"github.com/PuerkitoBio/goquery"
)

const thumbsDown = ":thumbsdown:"
const thumbsUp = ":thumbsup:"

var Headers = map[string]string{
	"authority":                 "www.supremecommunity.com",
	"cache-control":             "max-age=0",
	"upgrade-insecure-requests": "1",
	"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36",
	"sec-fetch-dest":            "document",
	"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	"sec-fetch-site":            "none",
	"sec-fetch-mode":            "navigate",
	"sec-fetch-user":            "?1",
	"accept-language":           "en-US,en;q=0.9",
}
var Client = &http.Client{}

type DropItem struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Price       Price  `json:"price,omitempty"`
	Votes       Votes  `json:"votes"`
	Link        string `json:"link,omitempty"`
}

type Price struct {
	FullPrice   string `json:"full_price"`
	DollarPrice string `json:"dollar_price"`
	PoundsPrice string `json:"pounds_price"`
}

type Votes struct {
	Upvotes   string `json:"upvotes"`
	Downvotes string `json:"downvotes"`
}

type Droplist []DropItem

func GetLatestDroplistLink() string {
	req, reqErr := http.NewRequest("GET", "https://www.supremecommunity.com/season/spring-summer2020/droplists/", nil)
	if reqErr != nil {
		fmt.Println(reqErr)
	}
	for k, v := range Headers {
		req.Header.Set(k, v)
	}
	res, resErr := Client.Do(req)
	if resErr != nil {
		fmt.Println(resErr)
	}

	doc, docErr := goquery.NewDocumentFromReader(res.Body)
	if docErr != nil {
		fmt.Println(docErr)
	}

	link, _ := doc.Find(".block").Attr("href")
	return link
}

func ScrapeDroplist(link string) Droplist {
	var list Droplist

	req, reqErr := http.NewRequest("GET", fmt.Sprintf("https://www.supremecommunity.com%s", link), nil)
	if reqErr != nil {
		fmt.Println(reqErr)
	}
	for k, v := range Headers {
		req.Header.Set(k, v)
	}
	res, resErr := Client.Do(req)
	defer res.Body.Close()
	if resErr != nil {
		fmt.Println(resErr)
	}
	doc, docErr := goquery.NewDocumentFromReader(res.Body)
	if docErr != nil {
		fmt.Println(docErr)
	}
	doc.Find(".masonry__item").Each(func(i int, s *goquery.Selection) {
		var Item DropItem

		Item.Name = s.Find("h2.name").Text()
		if len(Item.Name) < 2 {
			return
		}
		Item.Name = strings.Replace(Item.Name, "®", "", -1)

		img, imgExists := s.Find(".prefill-img").Attr("src")
		if imgExists {
			Item.Image = fmt.Sprintf("https://www.supremecommunity.com%s", img)
		}

		Item.Category = strings.Title(s.Find("p.category.hidden").Text())
		if len(Item.Category) < 2 {
			Item.Category = "N/A"
		}

		desc, descExists := s.Find(".prefill-img").Attr("alt")
		if descExists {
			if len(strings.Split(desc, "-")) > 1 {
				desc = strings.TrimSpace(strings.Split(desc, "-")[1])
			}
			Item.Description = strings.Replace(desc, "®", "", -1)
		}

		Item.Price.PoundsPrice = fmt.Sprintf("£%v", s.Find("p.pricegbp.hidden").Text())
		Item.Price.DollarPrice = fmt.Sprintf("$%v", s.Find("p.priceusd.hidden").Text())
		Item.Price.FullPrice = fmt.Sprintf("%s / %s", Item.Price.DollarPrice, Item.Price.PoundsPrice)
		Item.Votes.Upvotes = s.Find("div.progress-bar.progress-bar-success").Text()
		Item.Votes.Downvotes = s.Find("div.progress-bar.progress-bar-danger").Text()
		Item.Link = fmt.Sprintf("https://www.supremecommunity.com%s", link)

		list = append(list, Item)
	})
	return list
}

func SendDroplist(items Droplist, webhook string) {
	for _, item := range items {
		e := discord.NewEmbed(item.Name, item.Description, item.Link)
		e.SetThumbnail(item.Image)
		e.AddField("Price", item.Price.FullPrice, true)
		e.AddField("Category", item.Category, true)
		e.AddField("Votes", fmt.Sprintf("%v %v / %v %v", thumbsUp, item.Votes.Upvotes, thumbsDown, item.Votes.Downvotes), true)
		e.SetFooter("@aiomonitors", "")
		e.SetAuthor("Supreme Community", "https://twitter.com/Discoders", "")
		e.SendToWebhook(webhook)
		time.Sleep(500 * time.Millisecond)
	}
}
