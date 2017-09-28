package points

import (
	"fmt"
	"github.com/wavefronthq/go-proxy/common"
	"testing"
	"time"
)

func BenchmarkPointToStringBase(b *testing.B) {
	p := getPoint(1)
	for i := 0; i < b.N; i++ {
		pointToString(p)
	}
}

func BenchmarkPointToStringComplex(b *testing.B) {
	p := getPoint(10)
	for i := 0; i < b.N; i++ {
		pointToString(p)
	}
}

func getPoint(numTags int) *common.Point {
	point := &common.Point{}
	point.Name = "foo.metric.name"
	point.Value = "15.0"
	point.Source = "foo.source.name"
	point.Timestamp = time.Now().UnixNano() / 1e9

	point.Tags = make(map[string]string)
	for i := 1; i <= numTags; i++ {
		k, v := fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)
		point.Tags[k] = v
	}
	return point
}
