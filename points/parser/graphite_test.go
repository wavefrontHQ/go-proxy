package parser

import (
	"fmt"
	"testing"

	"github.com/wavefronthq/go-proxy/common"
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

	// exponential values
	"foo.metric 1e05 source=foo-linux",
	"foo.metric 1e05 1505454047 source=foo-linux",
	"foo.metric 1e-05 source=foo-linux",
	"foo.metric 1e-05 source=foo-linux",
	"foo.metric 1e5 1505454047 source=foo-linux",
	"foo.metric 1e-5 1505454047 source=foo-linux",

	// host
	"foo.metric 1.5 host=foo-linux",
	"foo.metric 1.5 1505454047 host=foo-linux",

	// verify extra space is allowed
	"foo.metric 1.5    1505454047    host=foo-linux",

	// tags
	"foo.metric 1.5 source=foo-linux env=dev",
	"foo.metric 1.5 source=foo-linux env=dev region=us-west2",
	"mac.disk.total 4.9895440384E11 1504118031 source=Vikrams-MacBook-Pro.local path=/ os=Mac device=disk1 fstype=hfs",

	// quotes
	"\"mac.disk.total\" 4.9895440384E11 1504118031 source=\"Vikrams-MacBook-Pro.local\" \"path\"=\"/\" \"os\"=\"Mac\" \"device\"=\"disk1\" \"fstype\"=\"hfs\"",
	"mac.cpu.usage.steal 0.000000 1505844752 cpu=\"cpu2\" os=\"Mac\" source=\"Vikrams-MacBook-Pro.local\"",

	// escaped quotes
	"foo.metric 1.5 source=foo-linux env=\"de\\\"v\"",
}

var invalidPoints = [...]string{
	"",
	"foo.metric",
	"foo.metric 1.5",
	"foo.metric 1",
	"foo.metric 1.5.0 source=foo-linux",
	"system.cpu.loadavg source=test.wavefront.com",
	"te\"st.metric 1 1505454047 source=test",
	"foo.metric 1e05e source=foo-linux",
}

func TestValidPoints(t *testing.T) {

	// Note: points valid for the parser may not be valid for the decoder (which performs additional validation)

	//TODO: add more test cases and flesh out the parser until all tests pass
	for _, pointLine := range validPoints {
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

func BenchmarkGraphiteParseBase(b *testing.B) {
	pt := "\"foo.metric\" 1.5 source=foo-linux \"env\"=\"dev\""
	for i := 0; i < b.N; i++ {
		graphiteParser.Parse([]byte(pt))
	}
}

func BenchmarkGraphiteParseComplex(b *testing.B) {
	pt := "\"mac.disk.total\" 4.9895440384E11 1504118031 source=\"Vikrams-MacBook-Pro.local\" \"path\"=\"/\" \"os\"=\"Mac\" \"device\"=\"disk1\" \"fstype\"=\"hfs\""
	for i := 0; i < b.N; i++ {
		graphiteParser.Parse([]byte(pt))
	}
}

func parsePoint(pt string) (*common.Point, error) {
	return graphiteParser.Parse([]byte(pt))
}
