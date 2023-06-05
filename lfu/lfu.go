package lfu

import (
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/secnot/orderedmap"
)

type (
	Node struct {
		lba  int
		op   string
		freq int
	}

	LFU struct {
		maxlen      int
		available   int
		hit         int
		miss        int
		pagefault   int
		writeCount  int
		readCount   int
		writeCost   float32
		readCost    float32
		eraseCost   float32
		maxfreq     int
		lastMinFreq int

		orderedList *orderedmap.OrderedMap
	}
)

func NewLFU(value int) *LFU {
	lfu := &LFU{
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
		maxfreq:     1000,
		lastMinFreq: 1,
		orderedList: orderedmap.NewOrderedMap(),
	}
	return lfu
}

func (lfu *LFU) evicted(data *Node) (minLBA int) {
	minFreqTemp := -1
	evictedLBA := 0
	iter := lfu.orderedList.Iter()
	for true {
		for key, _, ok := iter.Next(); ok; key, _, ok = iter.Next() {
			item, _ := lfu.orderedList.Get(key)
			node := item.(*Node)
			fmt.Println("[LOOPING DATA] : ", node.lba, node.op, node.freq)
			if lfu.lastMinFreq >= node.freq {
				evictedLBA = node.lba
				break
				// return node.lba
			}
			if minFreqTemp < node.freq {
				minFreqTemp = node.freq
			}
		}
		lfu.lastMinFreq = minFreqTemp
	}
	fmt.Println("ini adalah evictedLBA : ", evictedLBA)
	return evictedLBA
}

func (lfu *LFU) put(data *Node) (exists bool) {
	if _, _, ok := lfu.orderedList.GetLast(); !ok {
		fmt.Println("LFU cache is empty")
	}

	if item, ok := lfu.orderedList.Get(data.lba); ok {
		lfu.hit++

		node := item.(*Node)
		if node.freq < lfu.maxfreq {
			node.freq++
			fmt.Println(data.lba, data.op, node.freq)
		}

		if ok := lfu.orderedList.MoveLast(data.lba); !ok {
			fmt.Printf("Failed to move LBA %d to MRU position\n", data.lba)
		}
		return true
	} else {
		lfu.miss++
		lfu.readCount++
		if lfu.available > 0 {
			lfu.available--
			data.freq = 1
			lfu.orderedList.Set(data.lba, data)
		} else {
			lfu.pagefault++

			// Find item with minimum frequency
			minLBA := lfu.evicted(data)
			fmt.Println("tesmp : ", minLBA)

			// Remove item with minimum frequency
			lfuItem, _ := lfu.orderedList.Get(minLBA)
			lfuNode := lfuItem.(*Node)
			lfuOp := lfuNode.op
			if lfuOp == "W" {
				lfu.writeCount++
			}
			fmt.Println("[DELETED NODE] ", lfuNode.op, lfuNode.lba)

			lfu.orderedList.Delete(minLBA)

			data.freq = 1
			lfu.orderedList.Set(data.lba, data)
		}
		return false
	}
}

func (lfu *LFU) Get(trace simulator.Trace) (err error) {
	obj := new(Node)
	obj.lba = trace.Addr
	obj.op = trace.Op
	lfu.put(obj)

	return nil
}

func (lfu LFU) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	iter := lfu.orderedList.IterReverse()
	for _, value, ok := iter.Next(); ok; _, value, ok = iter.Next() {
		lruLba := value.(*Node)
		fmt.Println("[RESULT] : ", lruLba.lba, lruLba.op, lruLba.freq)
	}

	file.WriteString(fmt.Sprintf("cache size: %d\n", lfu.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lfu.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lfu.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", lfu.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", lfu.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float64(lfu.hit)/float64(lfu.hit+lfu.miss))*100))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n", float32(lfu.readCount)*lfu.readCost+float32(lfu.writeCount)*(lfu.writeCost+lfu.eraseCost)))
	file.WriteString(fmt.Sprintf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds()))
	return nil
}
