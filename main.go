package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/feeds"
	"net/http"
	"strconv"
	"strings"
	"time"
	"voltarss/data"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRss)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Starting server on :8080")

	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		panic(err)
	}
}

func getRss(ResponseWriter http.ResponseWriter, Request *http.Request) {
	page := Request.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	feed := &feeds.Feed{
		Title:       "reloia's - IISS VOLTA - DE GEMMIS",
		Description: "Feed RSS creato da ReLoia per l'IISS VOLTA - DE GEMMIS",
		Link:        &feeds.Link{Href: "https://iissvoltadegemmis.edu.it/circolare/", Rel: "self", Type: "text/html"},
		Author:      &feeds.Author{Name: "ReLoia", Email: "reloia@mntcrl.it"},
		Created:     time.Now(),
	}

	circolari := data.FetchCircolari(page)

	for _, c := range circolari {
		feed.Items = append(feed.Items, c.ToRssItem())
	}

	rss, err := feed.ToAtom()

	pageN, _ := strconv.Atoi(page)

	customLinks := "  <link rel=\"first\" href=\"https://reloia.ddns.net/voltarss?page=1\"></link>\n"
	if pageN > 1 {
		customLinks += fmt.Sprintf("  <link rel=\"prev\" href=\"https://reloia.ddns.net/voltarss?page=%s\"></link>\n", strconv.Itoa(pageN-1))
	}
	customLinks += fmt.Sprintf("  <link rel=\"next\" href=\"https://reloia.ddns.net/voltarss?page=%s\"></link>\n", strconv.Itoa(pageN+1))

	rss = strings.Replace(rss, "</subtitle>", "</subtitle>\n"+customLinks, 1)

	if err != nil {
		panic(err)
	}

	ResponseWriter.Header().Set("Content-Type", "application/rss+xml")
	ResponseWriter.WriteHeader(http.StatusOK)

	_, err = ResponseWriter.Write([]byte(rss))
}
