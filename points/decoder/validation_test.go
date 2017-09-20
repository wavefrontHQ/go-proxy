package decoder

import (
	"testing"
	"github.com/wavefronthq/go-proxy/common"
	"strings"
)

const VALID_NAME = "validName"
const VALID_SOURCE = "validSource"

func TestValidPoints(t *testing.T) {
	point := getPoint(VALID_NAME, VALID_SOURCE)
	err := validate(point)
	if err != nil {
		t.Error(err)
	}
}

func TestInvalidPoints(t *testing.T) {
	longName := getLongString(1025)
	point := getPoint(longName, VALID_SOURCE)
	err := validate(point)
	if err == nil {
		t.Error(err)
	}

	point = getPoint(VALID_NAME, longName)
	err = validate(point)
	if err == nil {
		t.Error(err)
	}
}

func getLongString(n int) string {
	return strings.Repeat("a", n)
}

func getPoint(name, source string) *common.Point {
	point := &common.Point{}
	point.Name = name
	point.Source = source
	return point
}


