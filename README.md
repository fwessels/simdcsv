# simdcsv

**Experimental**

Investigate whether 2 stage design approach used by [simdjson-go](https://github.com/minio/simdjson-go) can also speed up CSV parsing.

```$ go test -v -bench=Unaligned
pkg: github.com/fwessels/simdcsv
BenchmarkFindMarksUnaligned-8             525680              1959 ns/op        1659.02 MB/s        3456 B/op          1 allocs/op
PASS
```