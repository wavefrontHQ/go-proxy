package points

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-proxy/api"
)

// Interface that forwards points to a Wavefront instance.
type PointForwarder interface {
	init()
	addPoint(point string)
	checkOverflow()
	incrementBlockedPoint()
	receivedPoints() int64
	blockedPoints() int64
	sentPoints() int64
	queuedPoints() int64
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
	f.pointsFlushTime = metrics.GetOrRegisterTimer("push."+f.prefix+".duration", nil)
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
	if len(points) > 0 {
		f.points = append(points, f.points...)
	}
	f.mtx.Unlock()
	f.checkOverflow()
}

func (f *DefaultPointForwarder) addPoint(point string) {
	f.pointsReceived.Inc(1)
	f.mtx.Lock()
	f.points = append(f.points, point)
	f.mtx.Unlock()
}

func (f *DefaultPointForwarder) checkOverflow() {
	ptsLength := len(f.points)
	if ptsLength > f.maxBufferSize {
		log.Printf("Too many pending points: %d. Draining to queue.", ptsLength)
		f.drainToQueue()
	}
}

func (f *DefaultPointForwarder) drainToQueue() {
	f.mtx.Lock()
	ptsLength := len(f.points)
	overflow := ptsLength - f.maxBufferSize
	if overflow > 0 {
		// provide headroom for arriving points
		trimIdx := min(overflow+f.maxFlushSize, ptsLength)
		pointsToQueue := f.points[:trimIdx]

		if trimIdx == ptsLength {
			f.points = nil
		} else {
			f.points = f.points[trimIdx:]
		}
		f.mtx.Unlock()
		f.pointsQueued.Inc(int64(len(pointsToQueue)))
		bufferQueue.queuePoints(pointsToQueue)
	} else {
		f.mtx.Unlock()
	}
}

func (f *DefaultPointForwarder) incrementBlockedPoint() {
	f.pointsBlocked.Inc(1)
}

func (f *DefaultPointForwarder) receivedPoints() int64 {
	return f.pointsReceived.Count()
}

func (f *DefaultPointForwarder) blockedPoints() int64 {
	return f.pointsBlocked.Count()
}

func (f *DefaultPointForwarder) sentPoints() int64 {
	return f.pointsSent.Count()
}

func (f *DefaultPointForwarder) queuedPoints() int64 {
	return f.pointsQueued.Count()
}

func (f *DefaultPointForwarder) post(points []string) {
	ptsLength := len(points)
	if ptsLength == 0 {
		return
	}

	pointLines := strings.Join(points, "\n")
	resp, err := f.api.PostData(f.workUnitId, f.dataFormat, pointLines)

	if err != nil || (resp.StatusCode == api.NotAcceptableStatusCode) {
		if err != nil {
			log.Printf("%s: error posting data: %v\n", f.name, err)
		}
		f.buffer(points)
		return
	}
	f.pointsSent.Inc(int64(ptsLength))
}
