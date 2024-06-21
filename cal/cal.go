package cal

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/apognu/gocal"
	"github.com/makew0rld/1rg-server/config"
)

var events struct {
	SecondFloor []gocal.Event
	GreenRoom   []gocal.Event
	PurpleRoom  []gocal.Event
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
