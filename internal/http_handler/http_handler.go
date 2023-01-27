package http_handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	m "CinemaCityBrowser/internal/model"
)

const (
	cinemasURL        string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/cinemas/with-event/until/3000-06-06?attr=&lang=pl_PL`
	datesURL          string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/dates/in-cinema/1097/until/3000-06-06?attr=&lang=pl_PL`
	eventsURLtemplate string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/film-events/in-cinema/%v/at-date/%v?attr=&lang=pl_PL`
)

type HttpHandlerInterface interface {
	FetchCinemas() []m.Cinema
	FetchDates() []string
	FetchEvents(cinema m.Cinema, date string) ([]m.Film, []m.Event)
}

type HttpHandler struct {
	httpClient *http.Client
}

func (hh *HttpHandler) Init(dt time.Duration) {
	hh.httpClient = &http.Client{Timeout: dt}
}

func (hh *HttpHandler) fetchUrl(url string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := hh.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func (hh *HttpHandler) FetchCinemas() []m.Cinema {
	body := hh.fetchUrl(cinemasURL)
	var resp m.CinemasResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	cinemas := resp.Body.Cinemas
	return cinemas
}

func (hh *HttpHandler) FetchDates() []string {
	body := hh.fetchUrl(datesURL)
	var resp m.DatesResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	cinemas := resp.Body.Dates
	return cinemas
}

func (hh *HttpHandler) FetchEvents(cinema m.Cinema, date string) ([]m.Film, []m.Event) {
	url := fmt.Sprintf(eventsURLtemplate, cinema.Id, date)
	log.Println(url)
	body := hh.fetchUrl(url)
	var resp m.EventsResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	return resp.Body.Films, resp.Body.Events
}
