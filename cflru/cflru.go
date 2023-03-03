package cflru

import (
	"fmt"
	"os"
	"time"

	"lruwsr/simulator"

	"github.com/secnot/orderedmap"
)

type (
	Node struct {
		lba        int
		op         string
		dirtypages bool
	}

	CFLRU struct {
		maxlen          int
		available       int
		hit             int
		miss            int
		pagefault       int
		writeCount      int
		readCount       int
		writeCost       float32
		readCost        float32
		eraseCost       float32
		constWindowSize float32
		totalDirtyPages int

		orderedList *orderedmap.OrderedMap
	}
)

func NewCFLRU(value int) *CFLRU {
	cf := &CFLRU{
		maxlen:          value,
		available:       value,
		hit:             0,
		miss:            0,
		pagefault:       0,
		writeCount:      0,
		readCount:       0,
		writeCost:       0.25,
		readCost:        0.025,
		eraseCost:       2,
		constWindowSize: 0.1,
		totalDirtyPages: 0,
		orderedList:     orderedmap.NewOrderedMap(),
	}
	return cf
}

// func (cf *CFLRU) countWindowSize() {
// 	nd := cf.totalDirtyPages
// 	nc := (cf.maxlen - cf.available) - cf.totalDirtyPages
// 	w := (cf.writeCost * float32(nd)) - (cf.readCost)
// 	for _, item := range cf.ordeSredList.Items() {
// 	}
// 	cf.windowSize = int(math.Ceil(benefit / cost))
// }

func (cf *CFLRU) cleanFirst(data *Node) {
	windowSize := cf.constWindowSize * float32(cf.maxlen)
	iter := cf.orderedList.Iter()
	count := 0
	for key, value, ok := iter.Next(); ok && count <= int(windowSize); key, value, ok = iter.Next() {
		lruLba := value.(*Node)
		// fmt.Println(key, lruLba.op, lruLba.dirtypages)
		if !lruLba.dirtypages {
			// fmt.Println("dihapus [clean pages] : ", key, lruLba.op, lruLba.dirtypages)
			cf.orderedList.Delete(key)
			return
		}
		count++
	}

	if firstKey, _, ok := cf.orderedList.GetFirst(); ok {
		// fmt.Println("dihapus [dirty pages] : ", firstKey)
		cf.orderedList.Delete(firstKey)
		cf.writeCount++
	}
}

func (cf *CFLRU) put(data *Node) (exists bool) {
	if _, _, ok := cf.orderedList.GetLast(); !ok {
		fmt.Println("LRU cache is empty")
	}

	if _, ok := cf.orderedList.Get(data.lba); ok {
		cf.hit++
		if ok := cf.orderedList.MoveLast(data.lba); !ok {
			fmt.Println("Update Gagal")
		}

		// fmt.Println("ketemu ", data.lba)
		if ok := cf.orderedList.MoveLast(data.lba); !ok {
			fmt.Printf("Failed to move LBA %d to MRU position\n", data.lba)
		}

		return true
	} else {
		cf.miss++
		cf.readCount++
		if data.op == "W" {
			data.dirtypages = true
			cf.totalDirtyPages++
		}

		node := &Node{
			op:         data.op,
			dirtypages: data.dirtypages,
		}

		if cf.available > 0 {
			cf.available--
			cf.orderedList.Set(data.lba, node)
			// fmt.Println("masuk : ", data.lba)
		} else {
			cf.pagefault++
			if _, firstValue, ok := cf.orderedList.GetFirst(); ok {
				lruLba := firstValue.(*Node)
				if !lruLba.dirtypages {
					// fmt.Println("dihapus [clean pages] : ", key, lruLba.op, lruLba.dirtypages)
					cf.orderedList.PopFirst()
				} else {
					cf.cleanFirst(data)
				}
			} else {
				fmt.Println("No elements found to remove")
			}
			cf.orderedList.Set(data.lba, node)
		}
		return false
	}
}

func (cf *CFLRU) Get(trace simulator.Trace) (err error) {
	obj := new(Node)
	obj.lba = trace.Addr
	obj.op = trace.Op
	cf.put(obj)

	return nil
}

func (cf CFLRU) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("cache size: %d\n", cf.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", cf.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", cf.miss))
	file.WriteString(fmt.Sprintf("write count: %d\n", cf.writeCount))
	file.WriteString(fmt.Sprintf("read count: %d\n", cf.readCount))
	file.WriteString(fmt.Sprintf("hit ratio: %8.4f\n", (float64(cf.hit)/float64(cf.hit+cf.miss))*100))
	file.WriteString(fmt.Sprintf("runtime: %8.4f\n", float32(cf.readCount)*cf.readCost+float32(cf.writeCount)*(cf.writeCost+cf.eraseCost)))
	file.WriteString(fmt.Sprintf("time execution: %8.4f\n\n", time.Since(timeStart).Seconds()))
	return nil
}
