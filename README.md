# simdcsv

**Experimental: do not use in production.**

A 2 stage design approach for speeding up CSV parsing (somewhat analoguous to [simdjson-go](https://github.com/minio/simdjson-go)).
    
## Design goals

- 1 GB/sec parsing performance for a single core
- support arbitrarily large data sets
- drop-in replacement for `encoding/csv`
- zero copy behaviour/memory efficient

## Two Stage Design

The design of `simdcsv` consists of two stages:
- stage 1: preprocess the CSV
- stage 2: parse CSV

Fundamentally `simdcsv` works on chunks of 64 bytes at a time which are loaded into a set of 2 YMM registers. Using AVX2 instructions the presence of characters such as separators, newline delimiters and quotes are detected and merged into a single 64-bit wide register.

# Stage 1: Preprocessing stage

The main job of the first stage is to a chunk of data for the presence of quoted fields. 

Within quoted fields some special processing is done for:
- double quotes (`""`): these are cleared in the corresponding mask in order to prevent the next step to treat it from being treated as either an opening or closing qoute.
- separator character: each separator character in a quoted field has its corresponding bit cleared so as to be skipped in the next step.

In addition all CVS data is scanned for carriage return characters followed by a newline character (`\r\n`):
- within quoted fields: the (`\r\n`) pair is marked as a position to be replaced with a single newline (`\n`) only during the next stage.
- everywhere else: the `\r` is marked to be a newline (`\n`) so that it will be treated as an empty line during the next stage (which are ignored).

In order to be able to detect double quotes and carriage return and newline pairs, for the last bit (63rd) it is necessary to look ahead to the next set of 64 bytes. So the first stage reads one chunk ahead and, upon finishing the current chunk, swaps the next chunk to the current chunk and loads the next set of 64 bytes to process.

Special care is taken to prevent errors such as segmentation violations in order to not (potentially) load beyond the end of the input buffer.

The result from the first stage is a set of three bit-masks:
- quote mask: mask indicating opening or closing quotes for quoted fields (excluding quote quotes)
- separator mask: mask for splitsing the row of CSV data into separate fields (excluding separator characters in quoted fields)
- carriage return mask: mask than indicate which carriage returns to treat as newlines

# Stage 2: Parsing stage

The second stage takes the (adjusted) bit masks from the first stage in order the work out the offsets of the individual fields into the originating buffer containing the CSV data.

To prevent needlessy copying (and allocating) strings out of buffer of CSV data, there is a single large slice to strings representing all  the columns for all the rows in the CSV data. Each string out of the columns slice points back into the appropriate starting position in the CSV data and the corresponding length for that particular field.

As the columns are parsed they are added to the same row (which is a slice of strings) until a delimiter in the form of an active bit in the newline mask is detected. The a new row is started to which the subsequent fields are added.

Note that empty rows are automatically skipped as well as multiple empty rows immediately following each other. Due to the fact that a carriage return character, immediately followed by a newline, is indicated from the first stage to be treated as a newline character, it both properly terminates the current row as well as preventing an empty row from being added (since `\r\n` is treated as two subsequent newlines (`\n\n`) but empty lines are filtered out).


Due to the nature of CSV files this is not trivial by itself as for instance delimiter symbols are allowed in quoted fields. As such it is not possible to determine with certainty where chunks may be broken up at without doing additional processing.

##  Performance compared to encoding/csv

![encoding-csv_vs_simdcsv-comparison](charts/encoding-csv_vs_simdcsv.png)

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

