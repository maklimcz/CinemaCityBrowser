package model

type WithIndex interface {
	GetId() string
}

type CinemasResponse struct {
	Body struct {
		Cinemas []Cinema
	}
}

type Cinema struct {
	Id   string `bson:"_id"`
	Name string `json:"displayName" bson:"name"`
}

func (c Cinema) GetId() string {
	return c.Id
}

type DatesResponse struct {
	Body struct {
		Dates []string
	}
}

type EventsResponse struct {
	Body struct {
		Films  []Film
		Events []Event
	}
}

type Event struct {
	Id            string   `bson:"_id"`
	FilmId        string   `bson:"filmId"`
	CinemaId      string   `bson:"cinemaId"`
	BusinessDay   string   `bson:"businessDay"`
	EventDateTime string   `bson:"eventDateTime"`
	Attributes    []string `json:"attributeIds"`
	BookingLink   string   `bson:"bookingLink"`
	Auditorium    string   `json:"AuditoriumTinyName"`
}

func (e Event) GetId() string {
	return e.Id
}

type Film struct {
	Id          string `bson:"_id"`
	Name        string
	Length      int
	PosterLink  string
	Link        string
	ReleaseYear string
}

func (f Film) GetId() string {
	return f.Id
}
