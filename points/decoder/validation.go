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
	charErrStr   = "Invalid character: %s"
)

var (
	ErrMissingSource = errors.New("Missing source tag")
)

func validate(point *common.Point) error {
	err := validateStr(point.Name, 1024)
	if err != nil {
		return err
	}

	err = validateStr(point.Source, 1024)
	if err != nil {
		return err
	}

	for k, v := range point.Tags {
		totalLen := len(k) + len(v)
		if totalLen >= 255 {
			return fmt.Errorf(lengthErrStr, 254, totalLen)
		}
		err = validateRunes(k)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateStr(s string, maxLen int) error {
	strLen := len(s)
	if strLen <= 0 || strLen >= maxLen {
		return fmt.Errorf(lengthErrStr, maxLen, strLen)
	}
	return validateRunes(s)
}

func validateRunes(s string) error {
	for idx, r := range s {
		// Legal characters are 44-57 (,-./ and numbers), 65-90 (upper), 97-122 (lower), 95 (_)
		if !(44 <= r && r <= 57) && !(65 <= r && r <= 90) && !(97 <= r && r <= 122) && r != 95 {
			if idx != 0 || r != 126 {
				// first character can be 126 (~)
				return fmt.Errorf(charErrStr, string(r))
			}
		}
	}
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
