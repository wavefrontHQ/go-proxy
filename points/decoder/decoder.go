package decoder

import (
	"errors"
	"strings"

	"github.com/wavefronthq/go-proxy/common"
	"github.com/wavefronthq/go-proxy/points/parser"
)

var (
	ErrInvalidPoint = errors.New("DecodeError: incorrect point format")
)

// Interface for decoding a point line
type PointDecoder interface {
	Decode(b []byte) (*common.Point, error)
}

type DefaultDecoder struct {
	parser *parser.PointParser
}

func (d *DefaultDecoder) Decode(b []byte) (*common.Point, error) {
	if b == nil {
		return &common.Point{}, ErrInvalidPoint
	}

	pointLine := string(b)
	pointLine = strings.TrimSpace(pointLine)
	if pointLine == "" {
		return &common.Point{}, ErrInvalidPoint
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
