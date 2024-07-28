package services

import (
	"fmt"
	"sync"
	"time"
)

const AREA_COUNT = 3

type WateringManual struct {
	On      bool
	Areas   [AREA_COUNT]bool
	AutoOff time.Duration
	Start   time.Time
}

type WateringInterval struct {
	Id       int
	On       bool
	Areas    [AREA_COUNT]bool
	Days     [7]bool
	Start    time.Duration
	Duration time.Duration
}

type WateringUpdate struct {
	item any
}

// Todo: Use DB
type Watering struct {
	mutex     sync.Mutex
	manual    WateringManual
	intervals []WateringInterval
	updates   chan WateringUpdate
	nextIId   int
}

func NewWatering() *Watering {
	w := Watering{updates: make(chan WateringUpdate, 8)}
	go w.UpdateManger()
	return &w
}

func (w *Watering) GetState() ([AREA_COUNT]bool, time.Duration) {
	n := time.Now()
	startOfDay := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
	timeOfDay := n.Sub(startOfDay)

	// At lest update for the next Day
	nextChange := time.Hour*24 - timeOfDay
	var areas [AREA_COUNT]bool

	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, wi := range w.intervals {
		if wi.On && wi.Days[n.Weekday()] {
			if timeOfDay < wi.Start {
				nextChange = min(nextChange, wi.Start-timeOfDay)
			} else if timeOfDay < wi.Start+wi.Duration {
				for i, area := range wi.Areas {
					if area {
						areas[i] = true
					}
				}
				nextChange = min(nextChange, wi.Start+wi.Duration-timeOfDay)
			}
		}
	}

	if w.manual.On && w.manual.AutoOff != 0 {
		diff := w.manual.Start.Add(w.manual.AutoOff).Sub(n)
		if diff <= 0 {
			w.manual.On = false
			w.updates <- WateringUpdate{item: "manual off"}
		} else {
			areas = w.manual.Areas
			nextChange = min(nextChange, diff)
		}
	}
	return areas, nextChange
}

func (w *Watering) UpdateManger() {
	timer := time.NewTimer(0)
	for {
		select {
		case t := <-timer.C:
			fmt.Println("Timer: ", t)
			state, nextChange := w.GetState()
			fmt.Println("Results: ", state, nextChange)
			timer.Reset(nextChange)
		case u := <-w.updates:
			fmt.Println("Update: ", u.item)
			state, nextChange := w.GetState()
			fmt.Println("Results: ", state, nextChange)
			timer.Reset(nextChange)
		}
	}
}

func (w *Watering) GetManual() WateringManual {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.manual
}

func (w *Watering) UpdateManual(on bool, areas [AREA_COUNT]bool, autoOff time.Duration) WateringManual {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if on && !w.manual.On {
		w.manual.Start = time.Now()
	}
	w.manual.On = on
	w.manual.Areas = areas
	w.manual.AutoOff = autoOff
	w.updates <- WateringUpdate{item: w.manual}
	return w.manual
}

func (w *Watering) CreateInterval() WateringInterval {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	wi := WateringInterval{Id: w.nextIId}
	w.nextIId += 1
	w.intervals = append(w.intervals, wi)
	w.updates <- WateringUpdate{item: wi}
	return wi
}

func (w *Watering) GetIntervals() []WateringInterval {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.intervals
}

func (w *Watering) GetInterval(id int) (WateringInterval, bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, interval := range w.intervals {
		if interval.Id == id {
			return interval, true
		}
	}
	return WateringInterval{}, false
}

func (w *Watering) UpdateInterval(wi WateringInterval) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, interval := range w.intervals {
		if interval.Id == wi.Id {
			w.intervals[i] = wi
			w.updates <- WateringUpdate{item: wi}
			return true
		}
	}
	return false
}

func (w *Watering) DeleteInterval(id int) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, interval := range w.intervals {
		if interval.Id == id {
			w.intervals = append(w.intervals[:i], w.intervals[i+1:]...)
			w.updates <- WateringUpdate{item: fmt.Sprint("Delete: ", id)}
			return true
		}
	}
	return false
}
