# simdcsv

**Experimental**

Investigate whether 2 stage design approach used by [simdjson-go](https://github.com/minio/simdjson-go) can also speed up CSV parsing.

```$ go test -v -bench=Unaligned
pkg: github.com/fwessels/simdcsv
BenchmarkFindMarksUnaligned-8             525680              1959 ns/op        1659.02 MB/s        3456 B/op          1 allocs/op
PASS
```

## References

Ge, Chang and Li, Yinan and Eilebrecht, Eric and Chandramouli, Badrish and Kossmann, Donald, [Speculative Distributed CSV Data Parsing for Big Data Analytics](https://www.microsoft.com/en-us/research/publication/speculative-distributed-csv-data-parsing-for-big-data-analytics/), SIGMOD 2019.

