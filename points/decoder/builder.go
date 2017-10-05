package decoder

import (
	"github.com/wavefronthq/go-proxy/points/parser"
)

var (
	graphiteElements = parser.NewGraphiteElements()
	openTSDBElements = parser.NewOpenTSDBElements()
)

type DecoderBuilder interface {
	Build() PointDecoder
}

type GraphiteBuilder struct{}
type OpenTSDBBuilder struct{}

func (GraphiteBuilder) Build() PointDecoder {
	decoder := &DefaultDecoder{}
	decoder.parser = &parser.PointParser{Elements: graphiteElements}
	return decoder
}

func (OpenTSDBBuilder) Build() PointDecoder {
	decoder := &DefaultDecoder{}
	decoder.parser = &parser.PointParser{Elements: openTSDBElements}
	return decoder
}
