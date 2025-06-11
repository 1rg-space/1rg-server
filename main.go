package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/makew0rld/1rg-server/cal"
	"github.com/makew0rld/1rg-server/config"
	"github.com/makew0rld/1rg-server/database"
	"github.com/makew0rld/1rg-server/rolodex"
	"github.com/makew0rld/1rg-server/templates"
)

//go:embed assets
var assets embed.FS

func main() {
	log.Print("starting")

	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = cal.LoadEvents()
	if err != nil {
		log.Fatal(err)
	}

	// Homepage handler
	http.HandleFunc("GET /", mainPageHandler)

	// Asset handler
	http.Handle("GET /assets/", http.FileServerFS(assets))

	// Module handlers
	rolodexHandler, err := rolodex.NewHandler(db)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("GET /rolodex/add", rolodexHandler.AddGetHandler)
	http.HandleFunc("POST /rolodex/add", rolodexHandler.AddPostHandler)

	log.Print("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// This handler was used as the default for an unhandled path
		http.NotFound(w, r)
		return
	}

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

	templates.RenderTemplate(w, "index", data)

}
