package decoder

import (
	"github.com/wavefronthq/go-proxy/common"
	"strings"
	"testing"
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
	handleExpectedError(t, point)

	point = getPoint(VALID_NAME, longName)
	handleExpectedError(t, point)

	point = getPoint("foo bar", VALID_SOURCE)
	handleExpectedError(t, point)

	point = getPoint("system.cpu.load#", VALID_SOURCE)
	handleExpectedError(t, point)

	point = getPoint("system.cpu.load\\", VALID_SOURCE)
	handleExpectedError(t, point)
}

func handleExpectedError(t *testing.T, point *common.Point) {
	err := validate(point)
	if err == nil {
		t.Errorf("Error expected but not detected for point: %v", point)
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
