package decoder

import (
	"errors"
	"fmt"

	"github.com/wavefronthq/go-proxy/common"
)

const (
	sourceKey    = "source"
	hostKey      = "host"
	lengthErrStr = "Expected length less than %d, found %d"
)

var (
	ErrMissingSource = errors.New("Missing source tag")
)

func validate(point *common.Point) error {
	nameLen := len(point.Name)
	if nameLen <= 0 || nameLen >= 1024 {
		return fmt.Errorf(lengthErrStr, 1024, nameLen)
	}

	sourceLen := len(point.Source)
	if sourceLen <= 0 || sourceLen >= 1024 {
		return fmt.Errorf(lengthErrStr, 1024, sourceLen)
	}

	for k, v := range point.Tags {
		totalLen := len(k) + len(v)
		if totalLen >= 255 {
			return fmt.Errorf(lengthErrStr, 254, totalLen)
		}
	}
	//TODO: validate source, metric name, tag key/value characters
	return nil
}

func handleSource(point *common.Point) error {
	if source, ok := point.Tags[sourceKey]; ok {
		delete(point.Tags, sourceKey)
		point.Source = source
		return nil
	} else {
		if host, ok := point.Tags[hostKey]; ok {
			delete(point.Tags, sourceKey)
			point.Source = host
			return nil
		}
	}
	return ErrMissingSource
}
