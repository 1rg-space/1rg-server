package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/makew0rld/1rg-server/cal"
	"github.com/makew0rld/1rg-server/config"
)

//go:embed assets
//go:embed templates
var content embed.FS

var templates = template.Must(template.ParseFS(content, "templates/*"))

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	log.Print("starting")

	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = cal.LoadEvents()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			MeetingRooms  string
			PublicEvents  []cal.SimpleEvent
			PrivateEvents string
		}{}

		sf, gr, pr := cal.SecondFloorBusy(), cal.GreenRoomBusy(), cal.PurpleRoomBusy()
		// Eight combos: 000 to 111
		if !sf && !gr && !pr {
			data.MeetingRooms = "No meeting rooms are currently booked."
		} else if !sf && !gr && pr {
			data.MeetingRooms = "The purple meeting room is booked, but the others are free."
		} else if !sf && gr && !pr {
			data.MeetingRooms = "The green meeting room is booked, but the others are free."
		} else if !sf && gr && pr {
			data.MeetingRooms = "Only the second floor meeting room is free."
		} else if sf && !gr && !pr {
			data.MeetingRooms = "The second floor meeting room is booked, but the others are free."
		} else if sf && !gr && pr {
			data.MeetingRooms = "Only the green meeting room is free."
		} else if sf && gr && !pr {
			data.MeetingRooms = "Only the purple meeting room is free."
		} else if sf && gr && pr {
			data.MeetingRooms = "All meeting rooms are booked."
		}

		data.PrivateEvents = "I'm not sure if there are any private events, try checking Discord."
		data.PublicEvents = cal.PublicEventsToday()

		renderTemplate(w, "index", data)
	})
	http.Handle("GET /assets/", http.FileServerFS(content))

	log.Print("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
