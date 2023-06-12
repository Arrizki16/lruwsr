package lru

import (
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/secnot/orderedmap"
)

var FLAG = 1

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
	// fmt.Println("---------- PUTARAN KE ", FLAG, "----------")
	// FLAG += 1
	// fmt.Println("write count : ", lru.writeCount)
	// fmt.Println("hit ratio : ", lru.hit)
	// fmt.Println("pagefault : ", lru.pagefault)

	// iter := lru.orderedList.IterReverse()
	// for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
	// 	lruLba := value.(*Node)
	// 	fmt.Println("[RESULT] : ", lruLba.lba, lruLba.op)
	// }

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
		lru.readCount++

		node := &Node{
			lba: data.lba,
			op:  data.op,
		}

		if lru.available > 0 {
			lru.available--
			lru.orderedList.Set(data.lba, node)
		} else {
			lru.pagefault++

			if _, firstValue, ok := lru.orderedList.GetFirst(); ok {
				lruLba := firstValue.(*Node)

				if lruLba.op == "W" {
					lru.writeCount++
				}
				// fmt.Println("data yang dihapus : ", lruLba.lba)
				lru.orderedList.PopFirst()
			} else {
				fmt.Println("No elements found to remove")
			}
			// fmt.Println("masuk udah full : ", data.lba)
			lru.orderedList.Set(data.lba, node)
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
	fmt.Printf("cache size: %d\n", lru.maxlen)
	// fmt.Printf("cache hit: %d\n", lru.hit)
	// fmt.Printf("cache miss: %d\n", lru.miss)
	fmt.Printf("write count: %d\n", lru.writeCount)
	// fmt.Printf("read count: %d\n", lru.readCount)
	fmt.Printf("hit ratio: %8.4f\n\n", (float64(lru.hit)/float64(lru.hit+lru.miss))*100)
	// fmt.Printf("runtime: %8.4f\n", float32(lru.readCount)*lru.readCost+float32(lru.writeCount)*(lru.writeCost+lru.eraseCost))
	// fmt.Printf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds())
	return nil
}
