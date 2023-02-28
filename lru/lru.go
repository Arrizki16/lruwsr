package lru

import (
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/secnot/orderedmap"
)

type (
	Node struct {
		lba int
		op  string
	}

	LRU struct {
		maxlen     int
		available  int
		hit        int
		miss       int
		pagefault  int
		writeCount int
		readCount  int
		writeCost  float32
		readCost   float32
		eraseCost  float32

		orderedList *orderedmap.OrderedMap
	}
)

func NewLRU(value int) *LRU {
	lru := &LRU{
		maxlen:      value,
		available:   value,
		hit:         0,
		miss:        0,
		pagefault:   0,
		writeCount:  0,
		readCount:   0,
		writeCost:   0.25,
		readCost:    0.025,
		eraseCost:   2,
		orderedList: orderedmap.NewOrderedMap(),
	}
	return lru
}

func (lru *LRU) put(data *Node) (exists bool) {
	if _, _, ok := lru.orderedList.GetLast(); !ok {
		fmt.Println("LRU cache is empty")
	}

	if _, ok := lru.orderedList.Get(data.lba); ok {
		lru.hit++

		if ok := lru.orderedList.MoveLast(data.lba); !ok {
			fmt.Printf("Failed to move LBA %d to MRU position\n", data.lba)
		}
		return true
	} else {
		lru.miss++
		if lru.available > 0 {
			if data.op == "R" {
				lru.readCount++
			}
			lru.available--
			lru.orderedList.Set(data.lba, data.op)
		} else {
			lru.pagefault++
			if data.op == "R" {
				lru.readCount++
			}

			if firstKey, firstValue, ok := lru.orderedList.GetFirst(); ok {
				lruLba := &Node{lba: firstKey.(int), op: firstValue.(string)}
				lruOp := lruLba.op
				// fmt.Println("ini adalah data terkahir : ", lruLba, lruOp, firstKey, firstValue)

				if lruOp == "W" {
					lru.writeCount++
				}
				lru.orderedList.Delete(firstKey)
			} else {
				fmt.Println("No elements found to remove")
			}

			lru.orderedList.Set(data.lba, data.op)
		}
		return false
	}
}

func (lru *LRU) Get(trace simulator.Trace) (err error) {
	obj := new(Node)
	obj.lba = trace.Addr
	obj.op = trace.Op
	lru.put(obj)

	return nil
}

func (lru LRU) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", lru.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", lru.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float64(lru.hit)/float64(lru.hit+lru.miss))*100))
	// file.WriteString(fmt.Sprintf("!LRU|%d|%d|%d\n", lru.maxlen, lru.hit, lru.writeCount))
	// file.WriteString(fmt.Sprintf("runtime: %f\n", time.Since(timeStart).Seconds()))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n\n", float32(lru.readCount)*lru.readCost+float32(lru.writeCount)*(lru.writeCost+lru.eraseCost)))
	return nil
}
