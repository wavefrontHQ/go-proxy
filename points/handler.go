package points

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/wavefronthq/go-proxy/api"
	"github.com/wavefronthq/go-proxy/common"
)

const (
	minForwarders    = 4
	maxForwarders    = 16
	minFlushInterval = 1000
)

// Interface that handles the reporting of points.
type PointHandler interface {
	init(numTasks, interval, buffer, maxFlush int, dataFormat, workUnitId string, service api.WavefrontAPI)
	stop()
	reportPoint(point *common.Point)
	reportPoints(points []*common.Point)
	handleBlockedPoint(pointLine string)
}

type DefaultPointHandler struct {
	name            string
	pointForwarders []PointForwarder
}

func (h *DefaultPointHandler) init(numForwarders, flushInterval, maxBufferSize, maxFlushSize int,
	dataFormat, workUnitId string, service api.WavefrontAPI) {

	if numForwarders <= 0 || numForwarders > maxForwarders {
		log.Printf("%s-handler: numForwarders=%d\n", h.name, numForwarders)
		numForwarders = minForwarders
	}

	if flushInterval < minFlushInterval {
		log.Printf("%s-handler: flushInterval=%d\n", h.name, flushInterval)
		flushInterval = minFlushInterval
	}

	h.pointForwarders = make([]PointForwarder, numForwarders)
	for i := 0; i < numForwarders; i++ {
		pointForwarder := &DefaultPointForwarder{
			name:          fmt.Sprintf("%s-forwarder-%d", h.name, i),
			prefix:        h.name,
			api:           service,
			dataFormat:    dataFormat,
			workUnitId:    workUnitId,
			maxFlushSize:  maxFlushSize,
			maxBufferSize: maxBufferSize,
			pushTicker:    time.NewTicker(time.Millisecond * time.Duration(flushInterval)),
		}
		h.pointForwarders[i] = pointForwarder
		pointForwarder.init()
	}
}

func (h *DefaultPointHandler) getForwarder() PointForwarder {
	index := rand.Intn(len(h.pointForwarders))
	return h.pointForwarders[index]
}

func (h *DefaultPointHandler) reportPoint(point *common.Point) {
	forwarder := h.getForwarder()
	forwarder.addPoint(pointToString(point))
}

func (h *DefaultPointHandler) reportPoints(points []*common.Point) {
	for _, point := range points {
		h.reportPoint(point)
	}
}

func (h *DefaultPointHandler) handleBlockedPoint(pointLine string) {
	log.Printf("%s-handler: blocked point: %s", h.name, pointLine)
	h.getForwarder().incrementBlockedPoint()
}

func (h *DefaultPointHandler) stop() {
	for _, forwarder := range h.pointForwarders {
		forwarder.stop()
	}
}

func pointToString(point *common.Point) string {
	//<metricName> <metricValue> <timestamp> source=<source> [pointTags]
	var buf bytes.Buffer
	buf.WriteString(strconv.Quote(point.Name))
	buf.WriteString(" ")
	buf.WriteString(point.Value)
	buf.WriteString(" ")
	buf.WriteString(strconv.FormatInt(point.Timestamp, 10))
	buf.WriteString(" source=")
	buf.WriteString(strconv.Quote(point.Source))

	for k, v := range point.Tags {
		buf.WriteString(" ")
		buf.WriteString(strconv.Quote(k))
		buf.WriteString("=")
		buf.WriteString(strconv.Quote(v))
	}
	return buf.String()
}
