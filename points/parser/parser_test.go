package parser

import (
	"fmt"
	"github.com/wavefronthq/go-proxy/common"
	"testing"
)

var graphiteParser = NewGraphiteParser()

var validPoints = [...]string{
	// floating point values
	"foo.metric 1.5 source=foo-linux",
	"foo.metric 0.0 source=foo-linux",
	"foo.metric 1.5 1505454047 source=foo-linux",

	// integer values
	"foo.metric 1 source=foo-linux",
	"foo.metric 1 1505454047 source=foo-linux",

	// host
	"foo.metric 1.5 host=foo-linux",
	"foo.metric 1.5 1505454047 host=foo-linux",

	// tags
	"foo.metric 1.5 source=foo-linux env=dev",
	"foo.metric 1.5 source=foo-linux env=dev region=us-west2",
	"mac.disk.total 4.9895440384E11 1504118031 source=Vikrams-MacBook-Pro.local path=/ os=Mac device=disk1 fstype=hfs",

	// quotes
	"\"mac.disk.total\" 4.9895440384E11 1504118031 source=\"Vikrams-MacBook-Pro.local\" \"path\"=\"/\" \"os\"=\"Mac\" \"device\"=\"disk1\" \"fstype\"=\"hfs\"",
	"mac.cpu.usage.steal 0.000000 1505844752 cpu=\"cpu2\" os=\"Mac\" source=\"Vikrams-MacBook-Pro.local\"",
}

var invalidPoints = [...]string{
	"",
	"foo.metric",
	"foo.metric 1.5",
	"foo.metric 1",
	"foo.metric 1.5.0 source=foo-linux",
	"system.cpu.loadavg source=test.wavefront.com",
}

func TestValidPoints(t *testing.T) {

	// Note: points valid for the parser may not be valid for the decoder (which performs additional validation)

	//TODO: add more test cases and flesh out the parser until all tests pass
	for _, pointLine := range validPoints {
		fmt.Println(pointLine)
		pt, err := parsePoint(pointLine)
		if err != nil {
			fmt.Println("Error", pointLine, err)
			t.Error(err)
		} else {
			err = validateSource(pt)
			if err != nil {
				fmt.Println("Source Error", pointLine, err)
				t.Error(err)
			}
		}
		//fmt.Println("Valid Point", pt)
	}
}

func validateSource(point *common.Point) error {
	source, sok := point.Tags["source"]
	host, hok := point.Tags["host"]

	if !sok && !hok {
		return fmt.Errorf("Missing source/host")
	}

	if sok && source == "" {
		return fmt.Errorf("Invalid source")
	}

	if hok && host == "" {
		return fmt.Errorf("Invalid host")
	}
	return nil
}

func TestInvalidPoints(t *testing.T) {
	for _, pointLine := range invalidPoints {
		pt, err := parsePoint(pointLine)
		if err == nil {
			// if no error check source tags
			err = validateSource(pt)
			if err == nil {
				t.Error(fmt.Errorf("Error expected but not detected"))
			}

		}
	}
}

func parsePoint(pt string) (*common.Point, error) {
	return graphiteParser.Parse([]byte(pt))
}
