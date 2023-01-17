package lruwsr

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
		writeSeq   int // menyimpan jumlah write pada sebuah page yang digunakan untuk mengurutkan page di dalam cache berdasarkan jumlah write
		elem       *list.Element
	}

	LRU struct {
		maxlen      int
		available   int
		totalaccess int
		hit         int
		miss        int
		pagefault   int
		write       int

		tlba    *llrb.LLRB
		lrulist *list.List
	}

	NodeLba Node
)

func (x *NodeLba) Less(than llrb.Item) bool {
	return x.lba < than.(*NodeLba).lba
}

func NewLRUWSR(cacheSize int) *LRU {
	lru := &LRU{
		maxlen:      cacheSize,
		available:   cacheSize,
		totalaccess: 0,
		hit:         0,
		miss:        0,
		pagefault:   0,
		lrulist:     list.New(),
		tlba:        llrb.New(),
	}
	return lru
}

// func (lru *LRU) reorder() {
// 	var el *list.Element
// 	for el = lru.lrulist.Front(); el != nil; el = el.Next() {
// 		node := el.Value.(*NodeLba)
// 		el := lru.lrulist.PushFront(node)
// 		node.elem = el
// 		lru.lrulist.MoveToFront(el)
// 	}
// }

func (lru *LRU) reorder() {
	var el *list.Element
	var writeSeqOne []*NodeLba
	for el = lru.lrulist.Front(); el != nil; el = el.Next() {
		node := el.Value.(*NodeLba)
		if node.writeSeq == 1 {
			writeSeqOne = append(writeSeqOne, node)
			lru.lrulist.Remove(el)
		}
	}
	for _, node := range writeSeqOne {
		el := lru.lrulist.PushFront(node)
		node.elem = el
	}
}

func (lru *LRU) put(data *NodeLba) (exists bool) {
	var el *list.Element
	kk := new(NodeLba)

	node := lru.tlba.Get((*NodeLba)(data))
	if node != nil {
		lru.hit++
		dd := node.(*NodeLba) // shortcut saja
		if data.op == "W" {   // di lruwsr perlu ditulis jumlah operasi write
			lru.write++
			dd.writeSeq++
			lru.reorder()
		}
		lru.lrulist.Remove(dd.elem)
		el = lru.lrulist.PushFront(dd.elem.Value)
		dd.elem = el // update elem
		return true
	} else { // not exist
		lru.miss++
		if data.op == "W" {
			lru.write++
			data.writeSeq = 1
		}
		if lru.available > 0 {
			lru.available--
			el = lru.lrulist.PushFront(data)
			lru.tlba.InsertNoReplace(data)
			data.elem = el
		} else {
			lru.pagefault++

			// delete dulu
			el = lru.lrulist.Back()
			lba := el.Value.(*NodeLba).lba
			kk.lba = lba
			lru.tlba.Delete(kk) // hapus dah
			lru.lrulist.Remove(el)

			// masukkan lagi
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
	// file.WriteString(fmt.Sprintf("NUM ACCESS: %d\n", lru.totalaccess))
	// file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	// file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	// file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	// file.WriteString(fmt.Sprintf("cache page fault: %d\n", lru.pagefault))
	file.WriteString(fmt.Sprintf("NUM ACCESS: %d\n", lru.totalaccess))
	file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	file.WriteString(fmt.Sprintf("ssd write: %d\n", lru.write))
	file.WriteString(fmt.Sprintf("hit ratio : %8.4f\n", (float64(lru.hit)/float64(lru.totalaccess))*100))
	file.WriteString(fmt.Sprintf("tlba size : %d\n", lru.tlba.Len()))
	file.WriteString(fmt.Sprintf("list size : %d\n", lru.lrulist.Len()))

	file.WriteString(fmt.Sprintf("!LRUWSR|%d|%d|%d\n", lru.maxlen, lru.hit, lru.write))
	file.WriteString(fmt.Sprintf("total time: %f\n\n", time.Since(timeStart).Seconds()))

	return nil
}
