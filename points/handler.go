package points

import (
	"fmt"
	"github.com/wavefronthq/go-proxy/api"
	"github.com/wavefronthq/go-proxy/common"
	"log"
	"math/rand"
	"time"
)

const MIN_FORWARDERS = 2
const MAX_FORWARDERS = 16
const MIN_FLUSH_INTERVAL = 1000

type PointHandler interface {
	init(numTasks, interval, buffer, maxFlush int, dataFormat, workUnitId string, service api.WavefrontAPI)
	stop()
	ReportPoint(point *common.Point)
	ReportPoints(points []*common.Point)
	HandleBlockedPoint(pointLine string)
}

type DefaultPointHandler struct {
	name            string
	pointForwarders []PointForwarder
}

func (handler *DefaultPointHandler) init(numForwarders, flushInterval, maxBufferSize, maxFlushSize int,
	dataFormat, workUnitId string, service api.WavefrontAPI) {

	if numForwarders <= 0 || numForwarders > MAX_FORWARDERS {
		log.Printf("%s-handler: numForwarders=%d\n", handler.name, numForwarders)
		numForwarders = MIN_FORWARDERS
	}

	if flushInterval < MIN_FLUSH_INTERVAL {
		log.Printf("%s-handler: flushInterval=%d\n", handler.name, flushInterval)
		flushInterval = MIN_FLUSH_INTERVAL
	}

	handler.pointForwarders = make([]PointForwarder, numForwarders)
	for i := 0; i < numForwarders; i++ {
		forwarderName := fmt.Sprintf("%s-forwarder-%d", handler.name, i)
		pointForwarder := &DefaultPointForwarder{
			name:          forwarderName,
			prefix:        handler.name,
			api:           service,
			dataFormat:    dataFormat,
			workUnitId:    workUnitId,
			maxFlushSize:  maxFlushSize,
			maxBufferSize: maxBufferSize,
			pushTicker:    time.NewTicker(time.Millisecond * time.Duration(flushInterval)),
		}
		handler.pointForwarders[i] = pointForwarder
		pointForwarder.init()
	}
}

func (handler *DefaultPointHandler) getPointForwarder() PointForwarder {
	index := rand.Intn(len(handler.pointForwarders))
	return handler.pointForwarders[index]
}

func (handler *DefaultPointHandler) ReportPoint(point *common.Point) {
	//log.Printf("%s-handler: %+v\n", handler.name, point)
	forwarder := handler.getPointForwarder()
	forwarder.addPoint(pointToString(point))
}

func (handler *DefaultPointHandler) ReportPoints(points []*common.Point) {
	for _, point := range points {
		handler.ReportPoint(point)
	}
}

func (handler *DefaultPointHandler) HandleBlockedPoint(pointLine string) {
	log.Printf("%s-handler: blocked point: %s", handler.name, pointLine)
	handler.getPointForwarder().incrementBlockedPoint()
}

func pointToString(point *common.Point) string {
	//TODO: add quotes etc if not present around the point
	// look into inbuilt Quote function
	//<metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
	pointLine := fmt.Sprintf("%s %s %d source=%s", point.Name, point.Value, point.Timestamp, point.Source)
	for k, v := range point.Tags {
		pointLine = fmt.Sprintf(pointLine+" %s=%s", k, v)
	}
	return pointLine
}

func (handler *DefaultPointHandler) stop() {
	for _, forwarder := range handler.pointForwarders {
		forwarder.stop()
	}
}
