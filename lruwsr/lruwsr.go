package lruwsr

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
		coldFlag    bool
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
		decayPeriod    float32
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
		decayPeriod:    2,
		orderedList:    orderedmap.NewOrderedMap(),
	}
	return lru
}

func (lru *LRUWSR) reorder(data *Node) {
	min := 999
	flag := 999
	for {
		iter := lru.orderedList.Iter()
		for key, value, ok := iter.Next(); ok; key, value, ok = iter.Next() {
			lruLba := value.(*Node)
			if !lruLba.dirtypages {
				lru.orderedList.Delete(key)
				return
			}
			if flag > lruLba.accessCount {
				flag = lruLba.accessCount
			}
			if lruLba.accessCount < lru.coldTreshold || lruLba.accessCount == min {
				if lruLba.coldFlag {
					lru.writeCount++
					lru.orderedList.Delete(key)
					return
				} else {
					lruLba.coldFlag = true
					lru.orderedList.MoveLast(key)
				}
			} else if lruLba.accessCount >= lru.coldTreshold {
				lruLba.coldFlag = true
				// lruLba.accessCount = lruLba.access	Count - 1
				lru.orderedList.MoveLast(key)
			}
		}
		min = flag
	}
}

func (lru *LRUWSR) decay(data *Node) {
	iter := lru.orderedList.IterReverse()
	for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
		lruLba := value.(*Node)
		if !lruLba.dirtypages {
			continue
		}
		lruLba.accessCount = lruLba.accessCount - 1
	}
	lru.writeOperation = 0
}

func (lru *LRUWSR) Put(data *Node) (exists bool) {
	if value, ok := lru.orderedList.Get(data.lba); ok {
		lru.hit++
		lruLba := value.(*Node)
		if lruLba.op == "W" {
			lru.writeOperation++
			lruLba.coldFlag = false
			if lruLba.accessCount == 0 {
				lruLba.accessCount = 1
			} else if lruLba.accessCount < lru.maxlen {
				lruLba.accessCount = lruLba.accessCount + 1
			}
		}

		if ok := lru.orderedList.MoveLast(lruLba.lba); !ok {
			fmt.Printf("Failed to move LBA %d to MRU position\n", data.lba)
		}

		if float32(lru.writeOperation) == lru.decayPeriod*float32(lru.maxlen) {
			lru.decay(lruLba)
		}

		return true
	} else {
		lru.miss++
		lru.readCount++
		if data.op == "W" {
			data.dirtypages = true
			data.accessCount = 1
			lru.writeOperation++
			data.coldFlag = false
		} else {
			data.coldFlag = true
		}

		node := &Node{
			lba:         data.lba,
			op:          data.op,
			dirtypages:  data.dirtypages,
			accessCount: data.accessCount,
			coldFlag:    data.coldFlag,
		}

		if lru.available > 0 {
			lru.available--
			lru.orderedList.Set(data.lba, node)
			if float32(lru.writeOperation) == lru.decayPeriod*float32(lru.maxlen) {
				lru.decay(node)
			}
		} else {
			lru.pagefault++
			if _, firstValue, ok := lru.orderedList.GetFirst(); ok {
				lruLba := firstValue.(*Node)
				if !lruLba.dirtypages {
					lru.orderedList.PopFirst()
				} else {
					lru.reorder(lruLba)
				}
			} else {
				fmt.Println("No elements found to remove")
			}

			lru.orderedList.Set(data.lba, node)
			if float32(lru.writeOperation) == lru.decayPeriod*float32(lru.maxlen) {
				lru.decay(node)
			}
		}
		return false
	}
}

func (lru *LRUWSR) Get(trace simulator.Trace) (err error) {
	obj := new(Node)
	obj.lba = trace.Addr
	obj.op = trace.Op
	lru.Put(obj)

	return nil
}

func (lru LRUWSR) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", lru.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", lru.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float32(lru.hit)/float32(lru.hit+lru.miss))*100))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n", float32(lru.readCount)*lru.readCost+float32(lru.writeCount)*(lru.writeCost+lru.eraseCost)))
	file.WriteString(fmt.Sprintf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds()))

	return nil
}
