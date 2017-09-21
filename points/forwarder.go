package points

import (
	"github.com/wavefronthq/go-proxy/api"
	"log"
	"strings"
	"sync"
	"time"
)

type PointForwarder interface {
	flushPoints()
	addPoint(point string)
	stop()
}

type DefaultPointForwarder struct {
	name          string
	workUnitId    string
	dataFormat    string
	points        []string
	maxBufferSize int
	maxFlushSize  int
	mtx           sync.Mutex
	api           api.WavefrontAPI
	pushTicker    *time.Ticker
}

func (forwarder *DefaultPointForwarder) flushPoints() {
	for range forwarder.pushTicker.C {
		forwarder.postPoints(forwarder.getPointsBatch())
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
	log.Printf("%s: current points: %d", forwarder.name, currLen)
	return batchPoints
}

func (forwarder *DefaultPointForwarder) bufferPoints(points []string) {
	forwarder.mtx.Lock()
	currentSize := len(forwarder.points)
	pointsSize := len(points)

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
	//TODO: do not append if length greater than max buffer size?
	forwarder.mtx.Lock()
	forwarder.points = append(forwarder.points, point)
	forwarder.mtx.Unlock()
}

func (forwarder *DefaultPointForwarder) postPoints(points []string) {
	if len(points) == 0 {
		return
	}

	pointLines := strings.Join(points, "\n")
	resp, err := forwarder.api.PostData(forwarder.workUnitId, forwarder.dataFormat, pointLines)
	if err != nil {
		log.Println("Error posting data", err)
		forwarder.bufferPoints(points)
		return
	}
	log.Println(forwarder.name, "PostData Response Status:", resp.StatusCode)
	if resp.StatusCode == api.NOT_ACCEPTABLE_STATUS_CODE {
		forwarder.bufferPoints(points)
	}
}
