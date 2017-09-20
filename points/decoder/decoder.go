package decoder

import (
	"fmt"
	"errors"
	"github.com/wavefronthq/go-proxy/common"
)

const SOURCE = "source"
const HOST = "host"
var DECODE_ERROR = errors.New("DecodeError: incorrect point format")

// Interface for decoding a point line
type PointDecoder interface {
	Decode(b []byte) (*common.Point, error)
}

func handleSource(point *common.Point) error {
	source, ok := point.Tags[SOURCE];
	if ok {
		delete(point.Tags, SOURCE);
		point.Source = source
		return nil
	} else {
		host, ok := point.Tags[HOST];
		if ok {
			delete(point.Tags, SOURCE);
			point.Source = host
			return nil
		}
	}
	return fmt.Errorf("Missing source tag")
}

