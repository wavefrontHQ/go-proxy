package points

import "log"

type PointQueue interface {
	queuePoints(points []string)
}

type DefaultPointQueue struct{}

var bufferQueue = DefaultPointQueue{}

func (DefaultPointQueue) queuePoints(points []string) {
	//TODO: dropping on the floor for now, enhance to buffer to disk and process the buffer queues
	log.Printf("Dropping points of length: %d", len(points))
}
