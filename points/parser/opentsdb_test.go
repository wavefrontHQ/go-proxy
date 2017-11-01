package parser

import (
	"fmt"
	"testing"

	"github.com/wavefronthq/go-proxy/common"
)

var openTSDBParser = NewOpenTSDBParser()

var validTSDBPoints = [...]string{
	// floating point values
	"put foo.metric 1505454047 1.5 source=foo-linux",

	// integer values
	"put foo.metric 1505454047 1 source=foo-linux",

	// host
	"put foo.metric 1505454047 1.5 host=foo-linux",

	// tags
	"put mac.disk.total 1504118031 4.9895440384E11 source=Vikrams-MacBook-Pro.local path=/ os=Mac device=disk1 fstype=hfs",

	// quotes
	"put \"mac.disk.total\" 1504118031 4.9895440384E11 source=\"Vikrams-MacBook-Pro.local\" \"path\"=\"/\" \"os\"=\"Mac\" \"device\"=\"disk1\" \"fstype\"=\"hfs\"",
	"put mac.cpu.usage.steal 1505844752 0.000000 cpu=\"cpu2\" os=\"Mac\" source=\"Vikrams-MacBook-Pro.local\"",
}

var invalidTSDBPoints = [...]string{
	"",
	"foo.metric",
	"foo.metric 1.5",
	"foo.metric 1",
	"foo.metric 1.5.0 source=foo-linux",
	"system.cpu.loadavg source=test.wavefront.com",
	"put foo.metric 1 source=foo-linux",
	"put foo.metric 1.5 1505454047 host=foo-linux",
	"put foo.metric 1 1505454047 host=foo-linux",
}

func TestValidOpenTSDBPoints(t *testing.T) {

	// Note: points valid for the parser may not be valid for the decoder (which performs additional validation)

	//TODO: add more test cases and flesh out the parser until all tests pass
	for _, pointLine := range validTSDBPoints {
		pt, err := parseOpenTSDBPoint(pointLine)
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

func TestInvalidOpenTSDBPoints(t *testing.T) {
	for _, pointLine := range invalidTSDBPoints {
		pt, err := parseOpenTSDBPoint(pointLine)
		if err == nil {
			// if no error check source tags
			err = validateSource(pt)
			if err == nil {
				t.Error(fmt.Errorf("Error expected but not detected"))
			}

		}
	}
}

func BenchmarkOpenTSDBParseBase(b *testing.B) {
	pt := "\"foo.metric\" 1.5 source=foo-linux \"env\"=\"dev\""
	for i := 0; i < b.N; i++ {
		openTSDBParser.Parse([]byte(pt))
	}
}

func BenchmarkOpenTSDBParseComplex(b *testing.B) {
	pt := "\"mac.disk.total\" 4.9895440384E11 1504118031 source=\"Vikrams-MacBook-Pro.local\" \"path\"=\"/\" \"os\"=\"Mac\" \"device\"=\"disk1\" \"fstype\"=\"hfs\""
	for i := 0; i < b.N; i++ {
		openTSDBParser.Parse([]byte(pt))
	}
}

func parseOpenTSDBPoint(pt string) (*common.Point, error) {
	return openTSDBParser.Parse([]byte(pt))
}
