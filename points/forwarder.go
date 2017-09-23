package points

import (
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-proxy/api"
	"log"
	"strings"
	"sync"
	"time"
	//TODO: remove
	//"os"
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

func (forwarder *DefaultPointForwarder) init() {
	forwarder.pointsReceived = metrics.GetOrRegisterCounter("points."+forwarder.prefix+".received", nil)
	forwarder.pointsBlocked = metrics.GetOrRegisterCounter("points."+forwarder.prefix+".blocked", nil)
	forwarder.pointsQueued = metrics.GetOrRegisterCounter("points."+forwarder.prefix+".queued", nil)
	forwarder.pointsSent = metrics.GetOrRegisterCounter("points."+forwarder.prefix+".sent", nil)
	forwarder.pointsFlushTime = metrics.GetOrRegisterTimer("flush."+forwarder.prefix+".duration", nil)
	go forwarder.flushPoints()

	//TODO: report to the Wavefront instance instead (using wavefront-reporter or directly)
	//go metrics.Log(metrics.DefaultRegistry, 5 * time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
}

func (forwarder *DefaultPointForwarder) flushPoints() {
	for range forwarder.pushTicker.C {
		forwarder.pointsFlushTime.Time(func() {
			forwarder.post(forwarder.getPointsBatch())
		})
	}
	log.Printf("%s: exiting flushPoints", forwarder.name)
}

func (forwarder *DefaultPointForwarder) stop() {
	forwarder.pushTicker.Stop()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (forwarder *DefaultPointForwarder) getPointsBatch() []string {
	forwarder.mtx.Lock()
	currLen := len(forwarder.points)
	batchSize := min(currLen, forwarder.maxFlushSize)
	batchPoints := forwarder.points[:batchSize]
	forwarder.points = forwarder.points[batchSize:currLen]
	forwarder.mtx.Unlock()
	return batchPoints
}

func (forwarder *DefaultPointForwarder) buffer(points []string) {
	forwarder.mtx.Lock()
	currentSize := len(forwarder.points)
	pointsSize := len(points)
	forwarder.pointsQueued.Inc(int64(pointsSize))

	// do not add more points than the buffer is configured for
	trimSize := currentSize + pointsSize - forwarder.maxBufferSize
	if trimSize > 0 {
		retainSize := pointsSize - trimSize
		if retainSize <= 0 {
			points = nil
		} else {
			points = points[:retainSize]
		}
	}

	if len(points) > 0 {
		forwarder.points = append(points, forwarder.points...)
	}
	forwarder.mtx.Unlock()
}

func (forwarder *DefaultPointForwarder) addPoint(point string) {
	forwarder.pointsReceived.Inc(1)
	//TODO: do not append if length greater than max buffer size?
	forwarder.mtx.Lock()
	forwarder.points = append(forwarder.points, point)
	forwarder.mtx.Unlock()
}

func (forwarder *DefaultPointForwarder) incrementBlockedPoint() {
	forwarder.pointsBlocked.Inc(1)
}

func (forwarder *DefaultPointForwarder) post(points []string) {
	ptsLength := len(points)
	if ptsLength == 0 {
		return
	}

	pointLines := strings.Join(points, "\n")
	resp, err := forwarder.api.PostData(forwarder.workUnitId, forwarder.dataFormat, pointLines)

	if err != nil || (resp.StatusCode == api.NOT_ACCEPTABLE_STATUS_CODE) {
		if err != nil {
			log.Printf("%s: error posting data: %v\n", forwarder.name, err)
		}
		forwarder.buffer(points)
		return
	}
	forwarder.pointsSent.Inc(int64(ptsLength))
}
