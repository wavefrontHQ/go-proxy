package decoder

import (
	"errors"
	"github.com/wavefronthq/go-proxy/common"
)

var DECODE_ERROR = errors.New("DecodeError: incorrect point format")

// Interface for decoding a point line
type PointDecoder interface {
	Decode(b []byte) (*common.Point, error)
}
