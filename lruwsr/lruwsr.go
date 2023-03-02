package lruwsr

/*
getLast : mendapatkan elemen yang baru masuk
*/

import (
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/secnot/orderedmap"
)

type (
	Node struct {
		lba         int
		op          string
		accessCount int
		dirtypages  bool
	}

	LRUWSR struct {
		maxlen       int
		available    int
		hit          int
		miss         int
		pagefault    int
		writeCount   int
		readCount    int
		writeCost    float32
		readCost     float32
		eraseCost    float32
		coldTreshold int

		orderedList *orderedmap.OrderedMap
	}
)

func NewLRUWSR(value int) *LRUWSR {
	lru := &LRUWSR{
		maxlen:       value,
		available:    value,
		hit:          0,
		miss:         0,
		pagefault:    0,
		writeCount:   0,
		readCount:    0,
		writeCost:    0.25,
		readCost:     0.025,
		eraseCost:    2,
		coldTreshold: 1,
		orderedList:  orderedmap.NewOrderedMap(),
	}
	return lru
}

func (lru *LRUWSR) reorder(data *Node) {
	iter := lru.orderedList.Iter()
	for key, value, ok := iter.Next(); ok; key, value, ok = iter.Next() {
		lruLba := value.(*Node)
		if !lruLba.dirtypages {
			continue
		} else {
			if lruLba.accessCount < lru.coldTreshold {
				lru.orderedList.Delete(key)
			} else if lruLba.accessCount >= lru.coldTreshold {
				lruLba.accessCount--
				lru.orderedList.MoveLast(key)
			}
			return
		}
	}
}

func (lru *LRUWSR) put(data *Node) (exists bool) {
	if _, _, ok := lru.orderedList.GetLast(); !ok {
		fmt.Println("LRU cache is empty")
	}

	if value, ok := lru.orderedList.Get(data.lba); ok {
		lru.hit++
		lruLba := value.(*Node)
		if lruLba.op == "W" {
			if lruLba.accessCount == 0 {
				lruLba.accessCount = 1
			} else if lruLba.accessCount < lru.maxlen {
				lruLba.accessCount++
			}
		}

		if ok := lru.orderedList.MoveLast(data.lba); !ok {
			fmt.Printf("Failed to move LBA %d to MRU position\n", data.lba)
		}
		return true
	} else {
		lru.miss++
		lru.readCount++
		if data.op == "W" {
			data.dirtypages = true
			data.accessCount = 1
		}

		node := &Node{
			op:          data.op,
			dirtypages:  data.dirtypages,
			accessCount: data.accessCount,
		}

		if lru.available > 0 {
			lru.available--
			lru.orderedList.Set(data.lba, node)
		} else {
			lru.pagefault++
			if _, firstValue, ok := lru.orderedList.GetFirst(); ok {
				lruLba := firstValue.(*Node)
				if !lruLba.dirtypages {
					lru.orderedList.PopFirst()
				} else {
					lru.writeCount++
					lru.reorder(data)
				}
			} else {
				fmt.Println("No elements found to remove")
			}

			lru.orderedList.Set(data.lba, node)
		}
		return false
	}
}

func (lru *LRUWSR) Get(trace simulator.Trace) (err error) {
	obj := new(Node)
	obj.lba = trace.Addr
	obj.op = trace.Op
	lru.put(obj)

	return nil
}

func (lru LRUWSR) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", lru.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", lru.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float64(lru.hit)/float64(lru.hit+lru.miss))*100))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n\n", float32(lru.readCount)*lru.readCost+float32(lru.writeCount)*(lru.writeCost+lru.eraseCost)))
	return nil
}
