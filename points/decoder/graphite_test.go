package decoder

import (
	"testing"
)

func BenchmarkDecodeBase(b *testing.B) {
	var builder DecoderBuilder = GraphiteBuilder{}
	decoder := builder.Build()
	p := []byte("\"foo.metric\" 1.5 source=foo-linux \"env\"=\"dev\"")

	b.SetBytes(int64(len(p)))
	for i := 0; i < b.N; i++ {
		decoder.Decode(p)
	}
}

func BenchmarkDecodeComplex(b *testing.B) {
	var builder DecoderBuilder = GraphiteBuilder{}
	decoder := builder.Build()
	p := []byte("\"mac.disk.total\" 4.9895440384E11 1504118031 source=\"Vikrams-MacBook-Pro.local\" \"path\"=\"/\" \"os\"=\"Mac\" \"device\"=\"disk1\" \"fstype\"=\"hfs\"")

	b.SetBytes(int64(len(p)))
	for i := 0; i < b.N; i++ {
		decoder.Decode(p)
	}
}
