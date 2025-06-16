package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"1rg-server/cal"
	"1rg-server/config"
	"1rg-server/database"
	"1rg-server/rolodex"
	"1rg-server/templates"
)

//go:embed assets
var assets embed.FS

var calendarsProvided bool

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./1rg-server path/to/config.toml")
		return
	}

	log.Print("starting")

	log.Print("loading config")
	err := config.LoadConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if config.IsProduction {
		log.Print("mode: production")
	} else {
		log.Print("mode: debug")
	}

	log.Print("initializing database")
	db, err := database.Init()
	if err != nil {
		log.Fatal(err)
	}

	calendarsProvided = config.CalendarsProvided()
	if calendarsProvided {
		log.Print("loading events from calendars")
		err = cal.LoadEvents()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Print("not all calendars were configured, skipping")
	}

	log.Print("setting up HTTP handlers")

	// Homepage handler
	http.HandleFunc("GET /", mainPageHandler)

	// Asset handler
	// Use embedded assets in prod and disk assets when debugging
	if config.IsProduction {
		http.Handle("GET /assets/", http.FileServerFS(assets))
	} else {
		http.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	}

	// Module handlers
	rolodexHandler, err := rolodex.NewHandler(db)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("GET /rolodex/add", rolodexHandler.AddGetHandler)
	http.HandleFunc("POST /rolodex/add", rolodexHandler.AddPostHandler)

	log.Print("listening on http://localhost:8080")
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

	if calendarsProvided {
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
	} else {
		data.MeetingRooms = "Meeting room status: unknown"
		data.PrivateEvents = "Private events: unknown"
	}

	templates.RenderTemplate(w, "index", data)

}
