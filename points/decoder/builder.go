package decoder

import (
	"github.com/wavefronthq/go-proxy/points/parser"
	"log"
)

var GRAPHITE_ELEMENTS = parser.NewGraphiteElements()

type DecoderBuilder interface {
	Build() PointDecoder
}

type GraphiteBuilder struct{}

func (GraphiteBuilder) Build() PointDecoder {
	log.Println("Building new GraphiteDecoder")
	decoder := &GraphiteDecoder{}
	decoder.parser = &parser.PointParser{Elements: GRAPHITE_ELEMENTS}
	return decoder
}
