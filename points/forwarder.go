package points

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-proxy/api"
)

type PointForwarder interface {
	init()
	flushPoints()
	addPoint(point string)
	incrementBlockedPoint()
	stop()
}

type DefaultPointForwarder struct {
	name            string
	prefix          string
	workUnitId      string
	dataFormat      string
	points          []string
	maxBufferSize   int
	maxFlushSize    int
	mtx             sync.Mutex
	api             api.WavefrontAPI
	pushTicker      *time.Ticker
	pointsReceived  metrics.Counter
	pointsBlocked   metrics.Counter
	pointsQueued    metrics.Counter
	pointsSent      metrics.Counter
	pointsFlushTime metrics.Timer
}

func (f *DefaultPointForwarder) init() {
	f.pointsReceived = metrics.GetOrRegisterCounter("points."+f.prefix+".received", nil)
	f.pointsBlocked = metrics.GetOrRegisterCounter("points."+f.prefix+".blocked", nil)
	f.pointsQueued = metrics.GetOrRegisterCounter("points."+f.prefix+".queued", nil)
	f.pointsSent = metrics.GetOrRegisterCounter("points."+f.prefix+".sent", nil)
	f.pointsFlushTime = metrics.GetOrRegisterTimer("flush."+f.prefix+".duration", nil)
	go f.flushPoints()
}

func (f *DefaultPointForwarder) flushPoints() {
	for range f.pushTicker.C {
		f.pointsFlushTime.Time(func() {
			f.post(f.getPointsBatch())
		})
	}
	log.Printf("%s: exiting flushPoints", f.name)
}

func (f *DefaultPointForwarder) stop() {
	f.pushTicker.Stop()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (f *DefaultPointForwarder) getPointsBatch() []string {
	f.mtx.Lock()
	currLen := len(f.points)
	batchSize := min(currLen, f.maxFlushSize)
	batchPoints := f.points[:batchSize]
	f.points = f.points[batchSize:currLen]
	f.mtx.Unlock()
	return batchPoints
}

func (f *DefaultPointForwarder) buffer(points []string) {
	f.mtx.Lock()
	currentSize := len(f.points)
	pointsSize := len(points)
	f.pointsQueued.Inc(int64(pointsSize))

	// do not add more points than the buffer is configured for
	trimSize := currentSize + pointsSize - f.maxBufferSize
	if trimSize > 0 {
		retainSize := pointsSize - trimSize
		if retainSize <= 0 {
			points = nil
		} else {
			points = points[:retainSize]
		}
	}

	if len(points) > 0 {
		f.points = append(points, f.points...)
	}
	f.mtx.Unlock()
}

func (f *DefaultPointForwarder) addPoint(point string) {
	f.pointsReceived.Inc(1)
	//TODO: do not append if length greater than max buffer size?
	f.mtx.Lock()
	f.points = append(f.points, point)
	f.mtx.Unlock()
}

func (f *DefaultPointForwarder) incrementBlockedPoint() {
	f.pointsBlocked.Inc(1)
}

func (f *DefaultPointForwarder) post(points []string) {
	ptsLength := len(points)
	if ptsLength == 0 {
		return
	}

	pointLines := strings.Join(points, "\n")
	resp, err := f.api.PostData(f.workUnitId, f.dataFormat, pointLines)

	if err != nil || (resp.StatusCode == api.NOT_ACCEPTABLE_STATUS_CODE) {
		if err != nil {
			log.Printf("%s: error posting data: %v\n", f.name, err)
		}
		f.buffer(points)
		return
	}
	f.pointsSent.Inc(int64(ptsLength))
}
