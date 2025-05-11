package data

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"io"
	"strings"
	"time"
	"voltarss/network"
)

type Circolare struct {
	Titolo      string
	Descrizione string

	Link    string
	PdfLink string

	DataCreazione time.Time
}

func (c Circolare) ToRssItem() *feeds.Item {
	return &feeds.Item{
		Id:          c.Link,
		Title:       c.Titolo,
		Description: c.Descrizione,
		Author:      &feeds.Author{Name: "modugnon@gmail.com (Amministratore sito)", Email: "modugnon@gmail.com"},

		Link:      &feeds.Link{Href: c.Link},
		Enclosure: &feeds.Enclosure{Url: c.PdfLink},

		Created: c.DataCreazione,
	}
}

func FetchCircolari() []Circolare {
	res, err := network.MakeRequest("https://iissvoltadegemmis.edu.it/circolare/")

	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	if res.StatusCode != 200 {
		panic("Error: " + res.Status)
	}

	var circolari []Circolare
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}

	doc.Find("a.presentation-card-link").Each(func(i int, s *goquery.Selection) {
		pageLink := s.AttrOr("href", "")
		card := s.Find("article.card-article")

		if pageLink == "" || card == nil {
			return
		}

		titolo := strings.TrimSpace(card.Find("article.card-article").Text())
		shortDescription := strings.TrimSpace(card.Find(".card-article-content p").Text())

		circolare := SimpleCircolare{
			Titolo:      titolo,
			Descrizione: shortDescription,
			Link:        pageLink,
		}
		c := circolare.ToCircolare()
		circolari = append(circolari, *c)
	})

	return circolari
}
