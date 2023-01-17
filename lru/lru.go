package lru

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
		lba        int           // menyimpan alamat dari data yang disimpan dalam cache
		lastaccess int           // menyimpan waktu terakhir data dalam cache diakses
		op         string        // menyimpan operasi yang dilakukan pada data yang disimpan dalam cache
		elem       *list.Element // menyimpan posisi dari data yang disimpan dalam cache dalam linked list
	}

	LRU struct {
		maxlen      int // menyimpan ukuran maksimum dari cache
		available   int // menyimpan ukuran cache yang masih tersedia
		totalaccess int // menyimpan jumlah total akses yang dilakukan pada cache
		hit         int // persentase akses ke cache yang berhasil ditemukan di dalam cache
		miss        int // kebalikan dari hit
		pagefault   int // terjadi ketika cache sudah penuh dan harus mengosongkan cache untuk menambahkan data baru.
		write       int // menyimpan jumlah operasi write yang dilakukan pada cache

		tlba    *llrb.LLRB // menyimpan pointer ke LLRB (Left-Leaning Red-Black Tree) yang digunakan untuk menyimpan key-value dari data yang disimpan dalam cache
		lrulist *list.List // menyimpan pointer ke linked list yang digunakan untuk menyimpan posisi dari data yang disimpan dalam cache
	}

	NodeLba Node
)

func (x *NodeLba) Less(than llrb.Item) bool {
	return x.lba < than.(*NodeLba).lba
}

func NewLRU(cacheSize int) *LRU {
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

func (lru *LRU) put(data *NodeLba) (exists bool) {
	var el *list.Element
	kk := new(NodeLba) // untuk menyimpan data yang tidak ada di dalam cache

	node := lru.tlba.Get((*NodeLba)(data)) // library buat dapetin llrb.Item, jika data belum ada di cache maka nilai nil
	if node != nil {
		lru.hit++
		dd := node.(*NodeLba)
		if data.op == "W" {
			lru.write++
		}
		lru.lrulist.Remove(dd.elem)
		el = lru.lrulist.PushFront(dd.elem.Value)
		dd.elem = el
		return true
	} else {
		lru.miss++
		lru.write++
		if lru.available > 0 {
			lru.available--
			el = lru.lrulist.PushFront(data)
			lru.tlba.InsertNoReplace(data) // no replace karena datanya tidak ada
			data.elem = el                 // digunakan untuk menyimpan pointer ke elemen baru dalam linked list ke dalam struct data,
			//sehingga dapat digunakan untuk mengupdate posisi data dalam linked list nantinya.
		} else {
			lru.pagefault++
			el = lru.lrulist.Back()
			lba := el.Value.(*NodeLba).lba
			kk.lba = lba
			lru.tlba.Delete(kk)
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
	// {0, 'W'} -> address, operation
	obj.lba = trace.Addr // mendapatkan address dari trace
	obj.op = trace.Op    // mendapatkan operation dari trace
	obj.lastaccess = lru.totalaccess

	lru.put(obj)

	return nil
}

func (lru LRU) PrintToFile(file *os.File, timeStart time.Time) (err error) {
	file.WriteString(fmt.Sprintf("NUM ACCESS: %d\n", lru.totalaccess))
	file.WriteString(fmt.Sprintf("cache size: %d\n", lru.maxlen))
	file.WriteString(fmt.Sprintf("cache hit: %d\n", lru.hit))
	file.WriteString(fmt.Sprintf("cache miss: %d\n", lru.miss))
	file.WriteString(fmt.Sprintf("ssd write: %d\n", lru.write))
	file.WriteString(fmt.Sprintf("hit ratio : %8.4f\n", (float64(lru.hit)/float64(lru.totalaccess))*100))
	file.WriteString(fmt.Sprintf("tlba size : %d\n", lru.tlba.Len()))
	file.WriteString(fmt.Sprintf("list size : %d\n", lru.lrulist.Len()))

	file.WriteString(fmt.Sprintf("!LRU|%d|%d|%d\n", lru.maxlen, lru.hit, lru.write))
	file.WriteString(fmt.Sprintf("total time: %f\n\n", time.Since(timeStart).Seconds()))
	return nil
}
