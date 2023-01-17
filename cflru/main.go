package cflru

import (
	"container/list"
	"fmt"
	"lruwsr/simulator"
	"os"
	"time"

	"github.com/petar/GoLLRB/llrb"
)

type (
	Node struct {
		lba        int
		lastaccess int
		op         string
		clean      bool
		elem       *list.Element
	}

	CFLRU struct {
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

func NewCFLRU(cacheSize int) *CFLRU {
	cflru := &CFLRU{
		maxlen:      cacheSize,
		available:   cacheSize,
		totalaccess: 0,
		hit:         0,
		miss:        0,
		pagefault:   0,
		lrulist:     list.New(),
		tlba:        llrb.New(),
	}
	return cflru
}

func (cflru *CFLRU) put(data *NodeLba) (exists bool) {
	var el *list.Element
	kk := new(NodeLba)

	node := cflru.tlba.Get((*NodeLba)(data))
	if node != nil {
		cflru.hit++
		dd := node.(*NodeLba)
		if data.op == "W" {
			cflru.write++
			dd.clean = false
		}
		cflru.lrulist.Remove(dd.elem)
		el = cflru.lrulist.PushFront(dd.elem.Value)
		dd.elem = el
		return true
	} else {
		cflru.miss++
		cflru.write++
		data.clean = true //set clean flag to true
		if cflru.available > 0 {
			cflru.available--
			el = cflru.lrulist.PushFront(data)
			cflru.tlba.InsertNoReplace(data)
			data.elem = el
		} else {
			cflru.pagefault++
			// find the last clean page
			var lastClean *list.Element
			for el = cflru.lrulist.Back(); el != nil; el = el.Prev() {
				node := el.Value.(*NodeLba)
				if node.clean {
					lastClean = el
					break
				}
			}
			if lastClean != nil {
				// remove the clean page
				lba := lastClean.Value.(*NodeLba).lba
				kk.lba = lba
				cflru.tlba.Delete(kk)
				cflru.lrulist.Remove(lastClean)
			} else {
				// if there's no clean page found, remove the last page
				el = cflru.lrulist.Back()
				lba := el.Value.(*NodeLba).lba
				kk.lba = lba
				cflru.tlba.Delete(kk)
				cflru.lrulist.Remove(el)
			}
			// insert new page
			el = cflru.lrulist.PushFront(data)
			data.elem = el
			cflru.tlba.InsertNoReplace(data)
		}
		return false
	}
}

func (cflru *CFLRU) Get(trace simulator.Trace) (err error) {
	cflru.totalaccess++
	obj := new(NodeLba)
	obj.lba = trace.Addr
	obj.op = trace.Op
	obj.lastaccess = int(time.Now().UnixNano() / 1000000)
	cflru.put(obj)
	return
}

func (cflru CFLRU) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("NUM ACCESS: %d\n", cflru.totalaccess))
	file.WriteString(fmt.Sprintf("cache size: %d\n", cflru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", cflru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", cflru.miss))
	file.WriteString(fmt.Sprintf("ssd write: %d\n", cflru.write))
	file.WriteString(fmt.Sprintf("hit ratio : %8.4f\n", (float64(cflru.hit)/float64(cflru.totalaccess))*100))
	file.WriteString(fmt.Sprintf("tlba size : %d\n", cflru.tlba.Len()))
	file.WriteString(fmt.Sprintf("list size : %d\n", cflru.lrulist.Len()))

	file.WriteString(fmt.Sprintf("!LRUWSR|%d|%d|%d\n", cflru.maxlen, cflru.hit, cflru.write))
	file.WriteString(fmt.Sprintf("total time: %f\n\n", time.Since(timeStart).Seconds()))

	return nil
}
