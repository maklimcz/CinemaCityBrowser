package http_handler

import (
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"testing"

	m "CinemaCityBrowser/internal/model"
)

type MockHttpHandler struct{}

func (mhh MockHttpHandler) fetchUrl(url string) []byte {

	var fPath string
	switch {
	case strings.Contains(url, "dates"):
		fPath = "../../mockdata/dates.json"
	case strings.Contains(url, "cinemas"):
		fPath = "../../mockdata/cinemas.json"
	case strings.Contains(url, "events"):
		fPath = "../../mockdata/events.json"
	}

	content, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	return content
}

func TestFetchDates(t *testing.T) {
	var mhh MockHttpHandler

	dates := FetchDates(mhh)
	expectedDates := []string{
		"2022-12-22",
		"2022-12-23",
		"2022-12-25",
		"2022-12-26",
		"2022-12-27",
		"2022-12-28",
		"2022-12-29",
		"2022-12-30",
		"2022-12-31",
		"2023-01-01",
		"2023-01-02",
		"2023-01-03",
		"2023-01-04",
		"2023-01-05",
		"2023-01-27",
	}
	if !reflect.DeepEqual(dates, expectedDates) {
		t.Errorf("FetchDates returned %v, expected %v", dates, expectedDates)
	}
}

func TestFetchEvents(t *testing.T) {
	var mhh MockHttpHandler
	films, events := FetchEvents(mhh, m.Cinema{Id: "mockId", Name: "mockName"}, "mockDate")
	expectedFilmsNumber := 16
	expectedEventsNumber := 79

	if len(films) != expectedFilmsNumber {
		t.Errorf("FetchEvents returned %v films, expected %v", len(films), expectedFilmsNumber)
	}
	if len(events) != expectedEventsNumber {
		t.Errorf("FetchEvents returned %v events, expected %v", len(events), expectedEventsNumber)
	}
}

func TestFetchCinemas(t *testing.T) {
	var mhh MockHttpHandler
	cinemas := FetchCinemas(mhh)
	expectedCinemas := []m.Cinema{
		{Id: "1088", Name: "Bielsko-Biała"},
		{Id: "1086", Name: "Bydgoszcz"},
		{Id: "1092", Name: "Bytom"},
		{Id: "1098", Name: "Cieszyn"},
		{Id: "1089", Name: "Częstochowa - Galeria Jurajska"},
		{Id: "1075", Name: "Częstochowa - Wolność"},
		{Id: "1099", Name: "Elbląg"},
		{Id: "1085", Name: "Gliwice"},
		{Id: "1065", Name: "Katowice - Punkt 44"},
		{Id: "1079", Name: "Katowice - Silesia"},
		{Id: "1090", Name: "Kraków - Bonarka"},
		{Id: "1076", Name: "Kraków - Galeria Kazimierz"},
		{Id: "1064", Name: "Kraków - Zakopianka"},
		{Id: "1094", Name: "Lublin - Felicity"},
		{Id: "1084", Name: "Lublin - Plaza"},
		{Id: "1080", Name: "Łódź Manufaktura"},
		{Id: "1081", Name: "Poznań - Kinepolis"},
		{Id: "1078", Name: "Poznań - Plaza"},
		{Id: "1062", Name: "Ruda Śląska"},
		{Id: "1082", Name: "Rybnik"},
		{Id: "1083", Name: "Sosnowiec"},
		{Id: "1095", Name: "Starogard Gdański"},
		{Id: "1077", Name: "Toruń - Czerwona Droga"},
		{Id: "1093", Name: "Toruń - Plaza"},
		{Id: "1091", Name: "Wałbrzych"},
		{Id: "1074", Name: "Warszawa -  Arkadia"},
		{Id: "1061", Name: "Warszawa - Bemowo"},
		{Id: "1096", Name: "Warszawa - Białołęka Galeria Północna"},
		{Id: "1070", Name: "Warszawa - Galeria Mokotów"},
		{Id: "1069", Name: "Warszawa - Janki"},
		{Id: "1068", Name: "Warszawa - Promenada"},
		{Id: "1060", Name: "Warszawa - Sadyba"},
		{Id: "1067", Name: "Wrocław - Korona"},
		{Id: "1097", Name: "Wrocław - Wroclavia"},
		{Id: "1087", Name: "Zielona Góra"},
	}
	if !reflect.DeepEqual(cinemas, expectedCinemas) {
		t.Errorf("FetchCinemas returned %v, expected %v", cinemas, expectedCinemas)
	}
}
