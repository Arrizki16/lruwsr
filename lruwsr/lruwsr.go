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
		maxlen         int
		available      int
		hit            int
		miss           int
		pagefault      int
		writeCount     int
		readCount      int
		writeCost      float32
		readCost       float32
		eraseCost      float32
		coldTreshold   int
		writeOperation int
		decayPeriod    int
		orderedList    *orderedmap.OrderedMap
	}
)

func NewLRUWSR(value int) *LRUWSR {
	lru := &LRUWSR{
		maxlen:         value,
		available:      value,
		hit:            0,
		miss:           0,
		pagefault:      0,
		writeCount:     0,
		readCount:      0,
		writeCost:      0.25,
		readCost:       0.025,
		eraseCost:      2,
		coldTreshold:   1,
		writeOperation: 0,
		decayPeriod:    50000,
		orderedList:    orderedmap.NewOrderedMap(),
	}
	return lru
}

func (lru *LRUWSR) reorder(data *Node) {
	for {
		iter := lru.orderedList.Iter()
		for key, value, ok := iter.Next(); ok; key, value, ok = iter.Next() {
			lruLba := value.(*Node)
			if !lruLba.dirtypages {
				// fmt.Println("dihapus [clean pages] : ", key, lruLba.op, lruLba.dirtypages, lruLba.accessCount)
				lru.orderedList.Delete(key)
				return
			} else {
				if lruLba.accessCount < lru.coldTreshold {
					// fmt.Println("dihapus [dirty pages] : ", key, lruLba.op, lruLba.dirtypages, lruLba.accessCount)
					lru.writeCount++
					lru.orderedList.Delete(key)
					return
				} else if lruLba.accessCount >= lru.coldTreshold {
					// fmt.Println("movelast di reorder : ", key, lruLba.op, lruLba.dirtypages, lruLba.accessCount)
					lruLba.accessCount = 0
					lru.orderedList.MoveLast(key)
				}
			}
		}
	}
}

func (lru *LRUWSR) decay(data *Node) {
	// fmt.Println("write operation : ", lru.writeOperation)
	flag := 0
	iter := lru.orderedList.IterReverse()
	for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
		lruLba := value.(*Node)
		if !lruLba.dirtypages {
			continue
		} else {
			flag += 1
		}
		// fmt.Println("[FLAG] ", flag)
		// fmt.Println("print decay periods [Before]: ", key, lruLba.op, lruLba.dirtypages, lruLba.accessCount)
		lruLba.accessCount = lruLba.accessCount / 2
		// fmt.Println("print decay periods [After]: ", key, lruLba.op, lruLba.dirtypages, lruLba.accessCount)
	}
	flag = 0
	lru.writeOperation = 0
}

func (lru *LRUWSR) put(data *Node) (exists bool) {
	if _, _, ok := lru.orderedList.GetLast(); !ok {
		fmt.Println("LRU cache is empty")
	}

	if value, ok := lru.orderedList.Get(data.lba); ok {
		lru.hit++
		lruLba := value.(*Node)
		if lruLba.op == "W" {
			lru.writeOperation++
			if lruLba.accessCount == 0 {
				lruLba.accessCount = 1
			} else if lruLba.accessCount < lru.maxlen {
				lruLba.accessCount++
			}
		}

		// fmt.Println("ketemu ", data.lba, lruLba.accessCount)
		if ok := lru.orderedList.MoveLast(data.lba); !ok {
			fmt.Printf("Failed to move LBA %d to MRU position\n", data.lba)
		}

		if lru.writeOperation == lru.decayPeriod {
			lru.decay(data)
			// iter := lru.orderedList.IterReverse()
			// for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
			// 	lruLba := value.(*Node)
			// 	fmt.Println("[RESULT] : ", lruLba.lba, lruLba.op, lruLba.accessCount, lruLba.dirtypages)
			// }
			// fmt.Println("Jumlah write operation setelah decay : ", lru.writeOperation)
		}

		return true
	} else {
		lru.miss++
		lru.readCount++
		if data.op == "W" {
			data.dirtypages = true
			data.accessCount = 1
			lru.writeOperation++
		}

		node := &Node{
			lba:         data.lba,
			op:          data.op,
			dirtypages:  data.dirtypages,
			accessCount: data.accessCount,
		}

		if lru.available > 0 {
			lru.available--
			lru.orderedList.Set(data.lba, node)
			// fmt.Println("masuk : ", data.lba)
			if lru.writeOperation == lru.decayPeriod {
				lru.decay(data)
				// iter := lru.orderedList.IterReverse()
				// for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
				// 	lruLba := value.(*Node)
				// 	fmt.Println("[RESULT] : ", lruLba.lba, lruLba.op, lruLba.accessCount, lruLba.dirtypages)
				// }
			}
		} else {
			lru.pagefault++
			if _, firstValue, ok := lru.orderedList.GetFirst(); ok {
				lruLba := firstValue.(*Node)
				if !lruLba.dirtypages {
					// fmt.Println("dihapus [clean pages] : ", key, lruLba.op, lruLba.dirtypages, lruLba.accessCount)
					lru.orderedList.PopFirst()
				} else {
					lru.reorder(data)
				}
			} else {
				fmt.Println("No elements found to remove")
			}

			// fmt.Println("masuk udah full : ", data.lba)
			lru.orderedList.Set(data.lba, node)
			if lru.writeOperation == lru.decayPeriod {
				lru.decay(data)
			}
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
	// iter := lru.orderedList.IterReverse()
	// for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
	// 	lruLba := value.(*Node)
	// 	fmt.Println("[RESULT] : ", lruLba.lba, lruLba.op, lruLba.accessCount, lruLba.dirtypages)
	// }

	file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", lru.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", lru.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float64(lru.hit)/float64(lru.hit+lru.miss))*100))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n", float32(lru.readCount)*lru.readCost+float32(lru.writeCount)*(lru.writeCost+lru.eraseCost)))
	file.WriteString(fmt.Sprintf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds()))

	return nil
}
