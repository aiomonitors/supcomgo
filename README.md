# supcomgo

A simple [Supreme Community](https://supremecommunity.com) scraper written in GoLang, utilizing GoQuery, net/http, and a package written by myself that makes writing Discord embeds much easier.

#### Usage
```golang
package main

import "github.com/aiomonitors/supcomgo"

func main() {
    link := supcomgo.GetLatestDroplistLink()
    items := supcomgo.ScrapeDroplist(link)
    SendDroplist(items, webhook)
}```