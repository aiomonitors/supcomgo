package supcomgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	discord "github.com/aiomonitors/supcomgo/embeds"

	"github.com/PuerkitoBio/goquery"
)

const thumbsDown = ":thumbsdown:"
const thumbsUp = ":thumbsup:"

var headers = map[string]string{
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
var client = &http.Client{}

//DropItem is a struct for an item of the droplist
type DropItem struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Price       Price  `json:"price,omitempty"`
	Votes       Votes  `json:"votes"`
	Link        string `json:"link,omitempty"`
}

//Price is a struct for prices on an item
type Price struct {
	FullPrice   string `json:"full_price"`
	DollarPrice string `json:"dollar_price"`
	PoundsPrice string `json:"pounds_price"`
}

//Votes is a struct for votes on an item
type Votes struct {
	Upvotes   string `json:"upvotes"`
	Downvotes string `json:"downvotes"`
}

//Droplist is an array of DropItem's
type Droplist []DropItem

//GetLatestDroplistLink returns the latest droplist link on SupremeCommunity
func GetLatestDroplistLink() string {
	req, reqErr := http.NewRequest("GET", "https://www.supremecommunity.com/season/spring-summer2020/droplists/", nil)
	if reqErr != nil {
		fmt.Println(reqErr)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, resErr := client.Do(req)
	if resErr != nil {
		fmt.Println(resErr)
	}

	doc, docErr := goquery.NewDocumentFromReader(res.Body)
	if docErr != nil {
		fmt.Println(docErr)
	}

	link, _ := doc.Find(".block").Attr("href")
	return fmt.Sprintf("https://www.supremecommunity.com%s", link)
}

//ScrapeDroplist scrapes the droplsit from the link provided
func ScrapeDroplist(link string) Droplist {
	var list Droplist

	req, reqErr := http.NewRequest("GET", link, nil)
	if reqErr != nil {
		fmt.Println(reqErr)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, resErr := client.Do(req)
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
		Item.Link = link

		list = append(list, Item)
	})
	return list
}

//SendDroplist sends the droplist to a provided webhook
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

//ConvertToJSON converts a given droplist to JSON
//Returns []byte
func ConvertToJSON(items Droplist) ([]byte, error) {
	data, err := json.Marshal(items)
	return data, err
}

//SaveAsJSON saves a given Droplist item to a given path as JSON
//Returns error
func SaveAsJSON(items Droplist, path string) error {
	data, marshalError := json.Marshal(items)
	if marshalError != nil {
		return marshalError
	}
	saveError := ioutil.WriteFile(path, data, 0644)
	if saveError != nil {
		return saveError
	}
	return nil
}
