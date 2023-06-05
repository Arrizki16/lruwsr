# lruwsr
Repository Tugas Akhir

## Prerequirements
1. Install Go
2. Install Nim

## How To Run ?
1. Choose dataset
2. Compile converter ```split_financial.nim``` and ```split_websearch.num``` programs
```
nim c [nama program].nim
```
3. Make directory ```data``` in root directory

4. Running compile on dataset and choose ```data``` as directory target
```
./[nama program] [dataset] ../data/[output]
```
5. Go get module
```
go get github.com/petar/GoLLRB
go get github.com/secnot/orderedmap
```
6. Build ```main.go```
```
go build main.go
```
7. Running ```main``` and choose one algorithm
```
./main [algorithm(LRU|CFLRU|LRUWSR)] [file] [trace size]...
```