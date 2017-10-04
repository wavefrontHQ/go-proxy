package points

import (
	"fmt"
	"github.com/wavefronthq/go-proxy/api"
	"github.com/wavefronthq/go-proxy/common"
	"testing"
	"time"
)

func BenchmarkPointToStringBase(b *testing.B) {
	p := getPoint(1)
	h := &DefaultPointHandler{}
	h.init(2, 1000, 0, 0, "", "", &api.WavefrontAPIService{})
	for i := 0; i < b.N; i++ {
		h.pointToString(p)
	}
}

func BenchmarkPointToStringComplex(b *testing.B) {
	p := getPoint(10)
	h := &DefaultPointHandler{}
	h.init(2, 1000, 0, 0, "", "", &api.WavefrontAPIService{})
	for i := 0; i < b.N; i++ {
		h.pointToString(p)
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
