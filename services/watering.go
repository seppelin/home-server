package services

import (
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

// Todo: Use DB
type WaterinService struct {
	mutex     sync.Mutex
	manual    WateringManual
	intervals []WateringInterval
	nextIId   int
}

func NewWateringService() WaterinService {
	return WaterinService{}
}

func (ws *WaterinService) GetManual() WateringManual {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.manual
}

func (ws *WaterinService) UpdateManual(on bool, areas [3]bool, autoOff time.Duration) WateringManual {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if on && !ws.manual.On {
		ws.manual.Start = time.Now()
	}
	ws.manual.On = on
	ws.manual.Areas = areas
	ws.manual.AutoOff = autoOff
	return ws.manual
}

func (ws *WaterinService) CreateInterval() WateringInterval {
	wi := WateringInterval{Id: ws.nextIId}
	ws.nextIId += 1
	ws.intervals = append(ws.intervals, wi)
	return wi
}

func (ws *WaterinService) GetIntervals() []WateringInterval {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.intervals
}

func (ws *WaterinService) GetInterval(id int) (WateringInterval, bool) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	for _, interval := range ws.intervals {
		if interval.Id == id {
			return interval, true
		}
	}
	return WateringInterval{}, false
}

func (ws *WaterinService) UpdateInterval(wi WateringInterval) bool {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	for i, interval := range ws.intervals {
		if interval.Id == wi.Id {
			ws.intervals[i] = wi
			return true
		}
	}
	return false
}

func (ws *WaterinService) DeleteInterval(id int) bool {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	for i, interval := range ws.intervals {
		if interval.Id == id {
			ws.intervals = append(ws.intervals[:i], ws.intervals[i+1:]...)
			return true
		}
	}
	return false
}
