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

func (wi WateringInterval) GetId() string {
	return fmt.Sprintf("w-interval-%d", wi.Id)
}

type WateringUpdateKind int

const (
	UpdateManual WateringUpdateKind = iota
	CreateInterval
	UpdateInterval
	DeleteInterval
)

type WateringUpdate struct {
	ClientID string
	Kind     WateringUpdateKind
	Manual   WateringManual
	Interval WateringInterval
}

// Todo: Use DB
type Watering struct {
	mutex     sync.Mutex
	manual    WateringManual
	intervals []WateringInterval
	nextIId   int
	updates   chan WateringUpdate
}

func NewWatering() (*Watering, <-chan WateringUpdate, <-chan [AREA_COUNT]bool) {
	w := Watering{updates: make(chan WateringUpdate, 8)}
	web, ard := w.manager()
	return &w, web, ard
}

func (w *Watering) GetState() [AREA_COUNT]bool {
	n := time.Now()
	startOfDay := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
	timeOfDay := n.Sub(startOfDay)

	var areas [AREA_COUNT]bool

	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, wi := range w.intervals {
		if wi.On && wi.Days[n.Weekday()] {
			if timeOfDay < wi.Start+wi.Duration {
				for i, area := range wi.Areas {
					if area {
						areas[i] = true
					}
				}
			}
		}
	}
	if w.manual.On && w.manual.AutoOff != 0 {
		areas = w.manual.Areas
	}
	return areas
}

func (w *Watering) GetManual() WateringManual {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.manual
}

func (w *Watering) UpdateManual(on bool, areas [AREA_COUNT]bool, autoOff time.Duration, clientID string) WateringManual {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if on && (!w.manual.On || autoOff != 0) {
		w.manual.Start = time.Now()
	}
	w.manual.On = on
	w.manual.Areas = areas
	w.manual.AutoOff = autoOff
	w.updates <- WateringUpdate{ClientID: clientID, Kind: UpdateManual, Manual: w.manual}
	return w.manual
}

func (w *Watering) CreateInterval(clientID string) WateringInterval {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	wi := WateringInterval{Id: w.nextIId}
	w.nextIId += 1
	w.intervals = append(w.intervals, wi)
	w.updates <- WateringUpdate{ClientID: clientID, Kind: CreateInterval, Interval: wi}
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

func (w *Watering) UpdateInterval(wi WateringInterval, clientID string) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, interval := range w.intervals {
		if interval.Id == wi.Id {
			w.intervals[i] = wi
			w.updates <- WateringUpdate{ClientID: clientID, Kind: UpdateInterval, Interval: wi}
			return true
		}
	}
	return false
}

func (w *Watering) DeleteInterval(id int, clientID string) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, interval := range w.intervals {
		if interval.Id == id {
			w.intervals = append(w.intervals[:i], w.intervals[i+1:]...)
			w.updates <- WateringUpdate{ClientID: clientID, Kind: DeleteInterval, Interval: interval}
			return true
		}
	}
	return false
}

func (w *Watering) update(web chan<- WateringUpdate, ard chan<- [AREA_COUNT]bool, change *time.Timer) {
	n := time.Now()
	startOfDay := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
	timeOfDay := n.Sub(startOfDay)

	var areas [AREA_COUNT]bool
	// At lest update for the next Day
	nextDur := time.Hour*24 - timeOfDay

	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, wi := range w.intervals {
		if wi.On && wi.Days[n.Weekday()] {
			if timeOfDay < wi.Start {
				nextDur = min(nextDur, wi.Start-timeOfDay)
			} else if timeOfDay < wi.Start+wi.Duration {
				for i, area := range wi.Areas {
					if area {
						areas[i] = true
					}
				}
				nextDur = min(nextDur, wi.Start+wi.Duration-timeOfDay)
			}
		}
	}

	if w.manual.On && w.manual.AutoOff != 0 {
		diff := w.manual.Start.Add(w.manual.AutoOff).Sub(n)
		if diff <= 0 {
			w.manual.On = false
			web <- WateringUpdate{Kind: UpdateManual, Manual: w.manual}
		} else {
			areas = w.manual.Areas
			nextDur = min(nextDur, diff)
		}
	} else if w.manual.On {
		areas = w.manual.Areas
	}

	fmt.Println("UpdateFunc: ", areas, nextDur)
	ard <- areas
	change.Reset(nextDur)
}

func (w *Watering) manager() (<-chan WateringUpdate, <-chan [AREA_COUNT]bool) {
	web := make(chan WateringUpdate)
	ard := make(chan [AREA_COUNT]bool)
	change := time.NewTimer(0)

	go func() {
		for {
			select {
			case <-change.C:
				println("Change")
				w.update(web, ard, change)
			case u := <-w.updates:
				fmt.Println("Update: ", u.Kind)
				w.update(web, ard, change)
				web <- u
			}
		}
	}()

	return web, ard
}
