package data

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
	"voltarss/network"
)

type SimpleCircolare struct {
	Title       string
	Description string
	Link        string
}

type PdfElement struct {
	Title string
	Href  string
	Size  int
}

func parseSize(size string) int {
	size = strings.TrimSpace(size)
	size = strings.Trim(size, "()")
	size = strings.ToUpper(size)
	parts := strings.Split(size, " ")
	amount, unit := parts[0], parts[1]

	amountInt, err := strconv.Atoi(amount)

	if err != nil {
		panic(err)
	}

	var sizeInBytes int

	switch unit {
	case "KB":
		sizeInBytes = amountInt * 1024
	case "MB":
		sizeInBytes = amountInt * 1024 * 1024
	case "GB":
		sizeInBytes = amountInt * 1024 * 1024 * 1024
	default:
		panic("Unknown size unit: " + unit)
	}

	return sizeInBytes
}

func (c SimpleCircolare) ToCircolare() Circolare {
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
		fmt.Println("The document reader gave an error:", err)
		panic(err)
	}

	var pdfs []PdfElement

	pdfsEls := doc.Find(".icon.it-pdf-document").Parent()

	c.Description += "<br><h3>Documenti allegati</h3>"

	pdfsEls.Each(func(i int, s *goquery.Selection) {
		anchor := s.Find("a")
		size := s.Find("small").Text()
		size = strings.Split(size, "-")[1]
		size = strings.TrimSpace(size)

		pdf := PdfElement{
			Title: anchor.Text(),
			Href:  anchor.AttrOr("href", ""),
			Size:  parseSize(size),
		}

		pdfs = append(pdfs, pdf)

		c.Description += fmt.Sprintf("<a href=\"%s\">%s</a><br>", pdf.Href, pdf.Title)
	})

	sort.Slice(pdfs, func(i, j int) bool {
		return pdfs[i].Size > pdfs[j].Size
	})

	dateTxt := doc.Find("[data-element=\"metadata\"]").First().Text()
	dateTxt = strings.Split(strings.Split(dateTxt, "-")[0], ":")[1]
	dateTxt = strings.TrimSpace(dateTxt)
	dateTxt = strings.ReplaceAll(dateTxt, ".", "/")
	date, err := time.Parse("02/01/2006", dateTxt)
	if err != nil {
		panic(err)
	}

	return Circolare{
		Titolo:      c.Title,
		Descrizione: c.Description,
		Link:        c.Link,
		Pdf:         pdfs[0],

		DataCreazione: date,
	}
}
