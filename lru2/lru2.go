package lru2

import (
	"container/list"
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/petar/GoLLRB/llrb"
)

type (
	Node struct {
		lba        int
		lastaccess int
		op         string
		elem       *list.Element
	}

	LRU struct {
		totalaccess int
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

		tlba    *llrb.LLRB
		lrulist *list.List
	}

	NodeLba Node
)

func (x *NodeLba) Less(than llrb.Item) bool {
	return x.lba < than.(*NodeLba).lba
}

func NewLRU(value int) *LRU {
	lru := &LRU{
		totalaccess: 0,
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
		lrulist:     list.New(),
		tlba:        llrb.New(),
	}
	return lru
}

func (lru *LRU) put(data *NodeLba) (exists bool) {
	var el *list.Element
	kk := new(NodeLba)

	node := lru.tlba.Get((*NodeLba)(data))
	if node != nil {
		lru.hit++
		dd := node.(*NodeLba) // shortcut saja
		// if data.op == "W" {
		// lru.write++
		// }
		lru.lrulist.Remove(dd.elem)
		el = lru.lrulist.PushFront(dd.elem.Value)
		dd.elem = el // update elem
		return true
	} else { // not exist
		lru.miss++
		lru.readCount++
		if lru.available > 0 {
			lru.available--
			el = lru.lrulist.PushFront(data)
			lru.tlba.InsertNoReplace(data)
			data.elem = el
		} else {
			lru.pagefault++

			el = lru.lrulist.Back()
			lba := el.Value.(*NodeLba).lba
			op := el.Value.(*NodeLba).op
			if op == "W" {
				lru.writeCount++
			}
			kk.lba = lba
			lru.tlba.Delete(kk) // hapus dah
			lru.lrulist.Remove(el)

			el = lru.lrulist.PushFront(data)
			data.elem = el
			lru.tlba.InsertNoReplace(data)
		}
		return false
	}
}

func (lru *LRU) Get(trace simulator.Trace) (err error) {
	lru.totalaccess++
	obj := new(NodeLba)
	obj.lba = trace.Addr
	obj.op = trace.Op
	obj.lastaccess = lru.totalaccess

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
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n", float32(lru.readCount)*lru.readCost+float32(lru.writeCount)*(lru.writeCost+lru.eraseCost)))
	file.WriteString(fmt.Sprintf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds()))
	return nil
}
