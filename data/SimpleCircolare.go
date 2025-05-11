package data

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"time"
	"voltarss/network"
)

type SimpleCircolare struct {
	Titolo      string
	Descrizione string
	Link        string
}

func (c SimpleCircolare) ToCircolare() *Circolare {
	fmt.Println("Fetching circolare:", c.Link)

	res, err := network.MakeRequest(c.Link)
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

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}

	pdfs := doc.Find(".icon.it-pdf-document").Parent()

	fmt.Printf("Trovati %d pdfs\n", pdfs.Length())

	return &Circolare{
		Titolo:      c.Titolo,
		Descrizione: c.Descrizione,
		Link:        c.Link,
		PdfLink:     pdfs.AttrOr("href", ""),

		DataCreazione: time.Now(), // tODO: get the date from the page
	}
}
