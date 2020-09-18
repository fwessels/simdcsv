# simdcsv

**Experimental: do not use in production.**

A 2 stage design approach for speeding up CSV parsing (somewhat analoguous to [simdjson-go](https://github.com/minio/simdjson-go)).
    
## Design goals

- 1 GB/sec parsing performance for a single core
- linear performance scaling across cores
- support arbitrarily large data sets
- drop-in replacement for `encoding/csv`
- zero copy behaviour/memory efficient

## Two Stage Design

Fundamentally the architecture of `simdcsv` has two stages:
- stage 1: split up CSV
- stage 2: parse CSV

The first stage allows large CSV objects to be safely broken up into separate chunks that can be processed independently on multiple cores during the second stage. This is done in a deterministic manner whereby the "entry" state of each chunk is known definitively. 

Due to the nature of CSV files this is not trivial by itself as for instance delimiter symbols are allowed in quoted fields. As such it is not possible to determine with certainty where chunks may be broken up at without doing additional processing.

##  Performance compared to encoding/csv

```
benchmark                                     old MB/s     new MB/s     speedup
BenchmarkSimdCsv/parking-citations-100K-8     208.64       1178.09      5.65x
BenchmarkSimdCsv/worldcitiespop-8             127.65       1416.61      11.10x

benchmark                                     old bytes     new bytes     delta
BenchmarkSimdCsv/parking-citations-100K-8     58601190      1181503       -97.98%
BenchmarkSimdCsv/worldcitiespop-8             933054464     27603772      -97.04%
```

### Stage 1: preprocessing

### Stage 2: parse CSV

## Benchmarking 

### Stage 1

```
benchmark              old MB/s     new MB/s     speedup
BenchmarkFirstPass     760.36       4495.12      5.91x
```

```
go test -v -run=X -bench=Stage1PreprocessingMasks
BenchmarkStage1PreprocessingMasks-8       281197              3746 ns/op        1708.47 MB/s
```

### Stage 2

```
benchmark                        old MB/s     new MB/s     speedup
BenchmarkStage2ParseBuffer-8     205.81       1448.64      7.04x
```

Strongly reduced memory allocations:
```
benchmark                        old allocs     new allocs     delta
BenchmarkStage2ParseBuffer-8     20034          0              -100.00%
```

### Scaling across cores

```
$ go test -run=X -cpu=1,2,4,8,16 -bench=BenchmarkFirstPassAsm
BenchmarkFirstPassAsm              10000            109861 ns/op        4772.27 MB/s           0 B/op          0 allocs/op
BenchmarkFirstPassAsm-2            21762             55086 ns/op        9517.58 MB/s           0 B/op          0 allocs/op
BenchmarkFirstPassAsm-4            43603             27644 ns/op        18965.68 MB/s          0 B/op          0 allocs/op
BenchmarkFirstPassAsm-8            85539             13772 ns/op        38068.81 MB/s          0 B/op          0 allocs/op
BenchmarkFirstPassAsm-16          128840              9238 ns/op        56750.90 MB/s          0 B/op          0 allocs/op
```

## References

Ge, Chang and Li, Yinan and Eilebrecht, Eric and Chandramouli, Badrish and Kossmann, Donald, [Speculative Distributed CSV Data Parsing for Big Data Analytics](https://www.microsoft.com/en-us/research/publication/speculative-distributed-csv-data-parsing-for-big-data-analytics/), SIGMOD 2019.

[Awesome Comma-Separated Values](https://github.com/csvspecs/awesome-csv)

