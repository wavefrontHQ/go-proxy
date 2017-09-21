package decoder

import (
	"github.com/wavefronthq/go-proxy/points/parser"
)

var GRAPHITE_ELEMENTS = parser.NewGraphiteElements()

type DecoderBuilder interface {
	Build() PointDecoder
}

type GraphiteBuilder struct{}

func (GraphiteBuilder) Build() PointDecoder {
	decoder := &GraphiteDecoder{}
	decoder.parser = &parser.PointParser{Elements: GRAPHITE_ELEMENTS}
	return decoder
}
