package embeds

import (
	"strings"
	"strconv"
	"errors"
	"net/http"
	"encoding/json"
	"bytes"
)

type Embed struct {
	Username  string         `json:"username"`  
	AvatarURL string         `json:"avatar_url"`
	Content   string         `json:"content"`   
	Embeds    []EmbedElement `json:"embeds"`    
}

type EmbedElement struct {
	Author      Author  `json:"author"`     
	Title       string  `json:"title"`      
	URL         string  `json:"url"`        
	Description string  `json:"description"`
	Color       int64   `json:"color"`      
	Fields      []Field `json:"fields"`     
	Thumbnail   Image   `json:"thumbnail,omitempty"`  
	Image       Image   `json:"image,omitempty"`      
	Footer      Footer  `json:"footer"`     
}

type Author struct {
	Name    string `json:"name"`    
	URL     string `json:"url"`     
	IconURL string `json:"icon_url"`
}

type Field struct {
	Name   string `json:"name"`            
	Value  string `json:"value"`           
	Inline bool  `json:"inline,omitempty"`
}

type Footer struct {
	Text    string `json:"text"`    
	IconURL string `json:"icon_url,omitempty"`
}

type Image struct {
	URL string `json:"url"`
}

type Webhook struct {
	URL string `json:"url"`
	IconURL string `json:"icon_url"`
	Text string `json:"text"`
}

func NewEmbed(Title, Description, URL string) Embed {
	e := Embed{}
	emb := EmbedElement{Title: Title, Description: Description, URL: URL}
	e.Embeds  = append(e.Embeds, emb)
	return e
}

func (e *Embed) SetAuthor(Name, URL, IconURL string) {
	if len(e.Embeds) == 0 {
		emb := EmbedElement{Author: Author{Name, URL, IconURL}}
		e.Embeds = append(e.Embeds, emb)
	} else {
		e.Embeds[0].Author = Author{Name, URL, IconURL}
	}
}

func (e *Embed) SetColor(color string) error {
	color = strings.Replace(color, "0x", "", -1)
	color = strings.Replace(color, "0X", "", -1)
	colorInt, err := strconv.ParseInt(color, 16, 64)
	if err != nil {
		return errors.New("Invalid hex code passed")
	}
	e.Embeds[0].Color = colorInt
	return nil
}

func (e *Embed) SetThumbnail(URL string) error {
	if len(e.Embeds) < 1 {
		return errors.New("Invalid Embed passed in, Embed.Embeds must have at least one EmbedElement")
	}
	e.Embeds[0].Thumbnail = Image{URL}
	return nil
}

func (e *Embed) SetImage(URL string) error {
	if len(e.Embeds) < 1 {
		return errors.New("Invalid Embed passed in, Embed.Embeds must have at least one EmbedElement")
	}
	e.Embeds[0].Image = Image{URL}
	return nil
}

func (e *Embed) SetFooter(Text, IconURL string) error {
	if len(e.Embeds) < 1 {
		return errors.New("Invalid Embed passed in, Embed.Embeds must have at least one EmbedElement")
	}
	e.Embeds[0].Footer = Footer{Text, IconURL}
	return nil
}

func (e *Embed) AddField(Name, Value string, Inline bool) error {
	if len(e.Embeds) < 1 {
		return errors.New("Invalid Embed passed in, Embed.Embeds must have at least one EmbedElement")
	}
	e.Embeds[0].Fields = append(e.Embeds[0].Fields, Field{Name, Value, Inline})
	return nil
}

func (e *Embed) SendToWebhook(Webhook string) error {
	embed, marshalErr := json.Marshal(e); if marshalErr != nil {return marshalErr}
	_, postErr := http.Post(Webhook, "application/json", bytes.NewBuffer(embed))
	if postErr != nil {return postErr}
	return nil
}