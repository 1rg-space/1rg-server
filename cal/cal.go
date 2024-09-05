package cal

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/apognu/gocal"
	"github.com/makew0rld/1rg-server/config"
)

var events struct {
	SecondFloor  []gocal.Event
	GreenRoom    []gocal.Event
	PurpleRoom   []gocal.Event
	PublicEvents []gocal.Event
}
var eventsMu sync.RWMutex

func dayTrunc(t time.Time) time.Time {
	yy, mm, dd := t.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, t.Location())
}

func LoadEvents() error {
	eventsMu.Lock()
	defer eventsMu.Unlock()

	todayStart := dayTrunc(time.Now())
	todayEnd := dayTrunc(todayStart.AddDate(0, 0, 1))

	resp, err := http.Get(config.Config.Cals.SecondFloor)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	c := gocal.NewParser(resp.Body)
	c.Start = &todayStart
	c.End = &todayEnd
	err = c.Parse()
	if err != nil {
		return err
	}
	events.SecondFloor = c.Events
	resp.Body.Close()

	resp, err = http.Get(config.Config.Cals.GreenRoom)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	c = gocal.NewParser(resp.Body)
	c.Start = &todayStart
	c.End = &todayEnd
	err = c.Parse()
	if err != nil {
		return err
	}
	events.GreenRoom = c.Events
	resp.Body.Close()

	resp, err = http.Get(config.Config.Cals.PurpleRoom)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	c = gocal.NewParser(resp.Body)
	c.Start = &todayStart
	c.End = &todayEnd
	err = c.Parse()
	if err != nil {
		return err
	}
	events.PurpleRoom = c.Events
	resp.Body.Close()

	resp, err = http.Get(config.Config.Cals.PublicEvents)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	c = gocal.NewParser(resp.Body)
	c.Start = &todayStart
	c.End = &todayEnd
	err = c.Parse()
	if err != nil {
		return err
	}
	events.PublicEvents = c.Events
	resp.Body.Close()

	return nil
}

func SecondFloorBusy() bool {
	eventsMu.RLock()
	defer eventsMu.RUnlock()

	now := time.Now()
	for _, ev := range events.SecondFloor {
		if now.After(*ev.Start) && now.Before(*ev.End) {
			return true
		}
	}
	return false
}

func GreenRoomBusy() bool {
	eventsMu.RLock()
	defer eventsMu.RUnlock()

	now := time.Now()
	for _, ev := range events.GreenRoom {
		if now.After(*ev.Start) && now.Before(*ev.End) {
			return true
		}
	}
	return false
}

func PurpleRoomBusy() bool {
	eventsMu.RLock()
	defer eventsMu.RUnlock()

	now := time.Now()
	for _, ev := range events.PurpleRoom {
		if now.After(*ev.Start) && now.Before(*ev.End) {
			return true
		}
	}
	return false
}

type SimpleEvent struct {
	Name string
	Link string
}

var urlRe = regexp.MustCompile(`(?m)https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

func PublicEventsToday() []SimpleEvent {
	eventsMu.RLock()
	defer eventsMu.RUnlock()

	evts := make([]SimpleEvent, 0)
	for _, ev := range events.PublicEvents {
		evts = append(evts, SimpleEvent{
			Name: ev.Summary,
			Link: urlRe.FindString(ev.Description),
		})
	}
	return evts
}
