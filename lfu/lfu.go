package lfu

import (
	"container/list"
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/petar/GoLLRB/llrb"
)

const MAXFREQ = 1000

type (
	Node = struct {
		lba  int
		freq int
		op   string
		elem *list.Element
	}

	LFU struct {
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

		tlba    *llrb.LLRB
		freqArr [MAXFREQ]*list.List
	}
)

func NewLFU(value int) *LFU {
	lfu := &LFU{
		maxlen:     value,
		available:  value,
		hit:        0,
		miss:       0,
		pagefault:  0,
		writeCount: 0,
		readCount:  0,
		writeCost:  0.25,
		readCost:   0.025,
		eraseCost:  2,
		tlba:       llrb.New(),
		freqArr:    [MAXFREQ]*list.List{},
	}
	for i := 0; i < MAXFREQ; i++ {
		lfu.freqArr[i] = list.New()
	}
	return lfu
}

type NodeLba Node

func (x *NodeLba) Less(than llrb.Item) bool {
	return x.lba < than.(*NodeLba).lba
}

func (lfu *LFU) Put(data *NodeLba) (exists bool) {
	var el *list.Element
	kk := new(NodeLba)

	node := lfu.tlba.Get((*NodeLba)(data))
	if node != nil {
		lfu.hit++
		dd := node.(*NodeLba)
		if dd.freq < MAXFREQ {
			lst := lfu.freqArr[dd.freq-1]
			lst.Remove(dd.elem)
			dd.freq++
			lst = lfu.freqArr[dd.freq-1]
			el = lst.PushFront(dd.elem.Value)
			dd.elem = el
		}
		return true
	} else {
		lfu.miss++
		lfu.readCount++
		if lfu.available > 0 {
			lfu.available--
			el := lfu.freqArr[0].PushFront(data)
			data.elem = el
			lfu.tlba.InsertNoReplace(data)
		} else {
			lfu.pagefault++
			el = nil
			for ii := 0; ii < MAXFREQ; ii++ {
				if lfu.freqArr[ii].Len() > 0 {
					el = lfu.freqArr[ii].Back()
					lba := el.Value.(*NodeLba).lba
					op := el.Value.(*NodeLba).op
					if op == "W" {
						lfu.writeCount++
					}
					kk.lba = lba
					lfu.tlba.Delete(kk)
					lfu.freqArr[ii].Remove(el)
					break
				}
			}
			el = lfu.freqArr[0].PushFront(data)
			data.elem = el
			lfu.tlba.InsertNoReplace(data)
		}
		return false
	}
}

func (lfu *LFU) Get(trace simulator.Trace) (err error) {
	obj := new(NodeLba)
	obj.lba = trace.Addr
	obj.op = trace.Op
	obj.freq = 1
	lfu.Put(obj)

	return nil
}

func (lfu LFU) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("cache size: %d\n", lfu.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lfu.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lfu.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", lfu.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", lfu.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float32(lfu.hit)/float32(lfu.hit+lfu.miss))*100))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n", float32(lfu.readCount)*lfu.readCost+float32(lfu.writeCount)*(lfu.writeCost+lfu.eraseCost)))
	file.WriteString(fmt.Sprintf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds()))

	return nil
}
