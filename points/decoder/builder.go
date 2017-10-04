package decoder

import (
	"github.com/wavefronthq/go-proxy/points/parser"
)

var graphiteElements = parser.NewGraphiteElements()

type DecoderBuilder interface {
	Build() PointDecoder
}

type GraphiteBuilder struct{}

func (GraphiteBuilder) Build() PointDecoder {
	decoder := &GraphiteDecoder{}
	decoder.parser = &parser.PointParser{Elements: graphiteElements}
	return decoder
}
