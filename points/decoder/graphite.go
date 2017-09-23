package decoder

import (
	"strings"

	"github.com/wavefronthq/go-proxy/common"
	"github.com/wavefronthq/go-proxy/points/parser"
)

type GraphiteDecoder struct {
	parser *parser.PointParser
}

func (d *GraphiteDecoder) Decode(b []byte) (*common.Point, error) {
	if b == nil {
		return &common.Point{}, DECODE_ERROR
	}

	pointLine := string(b)
	pointLine = strings.TrimSpace(pointLine)
	if pointLine == "" {
		return &common.Point{}, DECODE_ERROR
	}

	point, err := d.parser.Parse(b)
	if err != nil {
		return point, err
	}
	err = handleSource(point)
	if err != nil {
		return point, err
	}
	return point, validate(point)
}
