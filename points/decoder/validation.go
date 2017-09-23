package decoder

import (
	"fmt"
	"github.com/wavefronthq/go-proxy/common"
)

const SOURCE = "source"
const HOST = "host"
const LENGTH_ERR = "Expected length less than %d, found %d"

func validate(point *common.Point) error {
	nameLen := len(point.Name)
	if nameLen <= 0 || nameLen >= 1024 {
		return fmt.Errorf(LENGTH_ERR, 1024, nameLen)
	}

	sourceLen := len(point.Source)
	if sourceLen <= 0 || sourceLen >= 1024 {
		return fmt.Errorf(LENGTH_ERR, 1024, nameLen)
	}

	for k, v := range point.Tags {
		totalLen := len(k) + len(v)
		if totalLen >= 255 {
			return fmt.Errorf(LENGTH_ERR, 254, totalLen)
		}
	}
	//TODO: validate source, metric name, tag key/value characters
	return nil
}

func handleSource(point *common.Point) error {
	if source, ok := point.Tags[SOURCE]; ok {
		delete(point.Tags, SOURCE)
		point.Source = source
		return nil
	} else {
		if host, ok := point.Tags[HOST]; ok {
			delete(point.Tags, SOURCE)
			point.Source = host
			return nil
		}
	}
	return fmt.Errorf("Missing source tag")
}
