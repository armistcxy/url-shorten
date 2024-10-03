package internal

import (
	"runtime"
	"testing"
)

func benchmarkGenerateRandomString(len int, b *testing.B) {
	var s string
	for range b.N {
		s = randomString(len)
	}
	runtime.KeepAlive(s)
}

func BenchmarkGenerateRandomStringLength4(b *testing.B) {
	benchmarkGenerateRandomString(4, b)
}

func BenchmarkGenerateRandomStringLength5(b *testing.B) {
	benchmarkGenerateRandomString(5, b)
}

func BenchmarkGenerateRandomStringLength6(b *testing.B) {
	benchmarkGenerateRandomString(6, b)
}

/*
go test -bench=Benchmark -benchmem
goos: linux
goarch: amd64
pkg: shorten/internal
cpu: AMD Ryzen 5 5600U with Radeon Graphics
BenchmarkGenerateRandomStringLength4-12           110550             10698 ns/op            5384 B/op          3 allocs/op
BenchmarkGenerateRandomStringLength5-12           109804             10716 ns/op            5386 B/op          3 allocs/op
BenchmarkGenerateRandomStringLength6-12           109509             10719 ns/op            5392 B/op          3 allocs/op
*/

// Base on the benchmark results above => There's not much different when generate between string length
// Just use Length = 6 => there are 62 ^ 6 possible URLs
