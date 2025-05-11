package data

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
	"voltarss/network"
)

var (
	cache      = sync.Map{}
	cacheTTL   = 2 * time.Hour
	cacheTimes = sync.Map{}
)

type Circolare struct {
	Titolo      string
	Descrizione string

	Link string
	Pdf  PdfElement

	DataCreazione time.Time
}

func (c Circolare) ToRssItem() *feeds.Item {
	return &feeds.Item{
		Id:          c.Link,
		Title:       c.Titolo,
		Description: c.Descrizione,
		Author:      &feeds.Author{Name: "modugnon@gmail.com (Amministratore sito)", Email: "modugnon@gmail.com"},

		Link:      &feeds.Link{Href: c.Link},
		Enclosure: &feeds.Enclosure{Url: c.Pdf.Href, Type: "application/pdf", Length: strconv.FormatInt(int64(c.Pdf.Size), 10)},

		Created: c.DataCreazione,
	}
}

func FetchCircolari(page string) []Circolare {
	number, _ := strconv.Atoi(page)

	sourcePage := strconv.Itoa((number + 1) / 2)
	isFirstFive := number%2 == 1

	cacheKey := "circolari_" + strconv.Itoa(number)
	if cached, found := cache.Load(cacheKey); found {
		value, _ := cacheTimes.LoadOrStore(cacheKey, time.Now())
		timee := value.(time.Time)

		if time.Since(timee) < cacheTTL {
			return cached.([]Circolare)
		}
	}

	res, err := network.MakeRequest("https://iissvoltadegemmis.edu.it/circolare/page/" + sourcePage)

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

	circEls := doc.Find("a.presentation-card-link")

	if isFirstFive {
		circEls = circEls.Slice(0, 5)
	} else {
		circEls = circEls.Slice(5, 10)
	}

	circEls.Each(func(i int, s *goquery.Selection) {
		pageLink := s.AttrOr("href", "")
		card := s.Find("article.card-article")

		if pageLink == "" || card == nil {
			return
		}

		titolo := strings.TrimSpace(card.Find("h2.h3").Text())
		shortDescription := strings.TrimSpace(card.Find(".card-article-content p").Text())

		circolare := SimpleCircolare{
			Title:       titolo,
			Description: shortDescription,
			Link:        pageLink,
		}
		c := circolare.ToCircolare()
		circolari = append(circolari, c)
	})

	cache.Store(cacheKey, circolari)
	cacheTimes.Store(cacheKey, time.Now())
	return circolari
}

func (c Circolare) ToString() string {
	return c.Titolo + "\n" + c.Descrizione + "\n" + c.Link
}
