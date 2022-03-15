package node

import (
	"io"
	"fmt"
	"sync"

	. "github.com/Vlad-Karna/vfs/vfuse"

	"github.com/billziss-gh/cgofuse/fuse"
)

type PagedFileInterface interface {
	io.Closer
	Size() int64
	PageSize() int
	ReadPage (b []byte, page int64) (n int, err error)
	WritePage(b []byte, page int64) (n int, err error)
}

type DynamicPagedFile struct {
	FileBase
	pf			PagedFileInterface	// Underlying PagedFile
	page			*dynamicPage	// Cached Page
	mutex			sync.Mutex	// Page Mutex
	lastpage		int64		// Last Page Number
	lastpagesize		uint		// Last Page Size
}

func NewDynamicPagedFile(pf PagedFileInterface) (res *DynamicPagedFile) {
	res = &DynamicPagedFile{ pf: pf}
	res.readSize()
	_, err := res.lockPage(0)
	if err != nil {
		res = nil
		return
	}
	res.unlockPage()
return
}

func (p *DynamicPagedFile) readSize() { // For mutable read-only files $$$
	sz := p.pf.Size()
	ps := int64(p.pf.PageSize())
	newlpg :=      sz / ps
	newlps := uint(sz % ps)
	if p.lastpage != newlpg || p.lastpagesize != newlps {
		p.lastpage     = newlpg
		p.lastpagesize = newlps
		p.freePage()
	}
}

func (p *DynamicPagedFile) Size() int64 {
	p.readSize()
return p.lastpage * int64(p.pf.PageSize()) + int64(p.lastpagesize)
}

func (p *DynamicPagedFile) Getattr(stat *fuse.Stat_t) (errc int) {
	errc = p.FileBase.Getattr(stat)
	stat.Size = p.Size()
return
}

func (p *DynamicPagedFile) ReadAt(b []byte, off int64) (n int, err error) {
	p.readSize()
	pgsz := int64(p.pf.PageSize())
	for len(b) > 0 {
		pn := off / pgsz
		pg, err := p.lockPage(pn)
		if err != nil {
			return n, err
		}
		rd := pg.readAt(b, uint(off % pgsz))
		p.unlockPage()
		if rd == 0 {
			break
		}
		off += int64(rd)
		n += rd
		b = b[rd:]
	}
return
}

func (p *DynamicPagedFile) Close() (err error) {
	p.freePage()
	err = p.pf.Close()
return
}

// Working with Pages

func (f *DynamicPagedFile) newPage() (p *dynamicPage) {
	p = new(dynamicPage)
	p.buf = make([]byte, f.pf.PageSize())
//	p.parent = f
return
}

func (f *DynamicPagedFile) flushPage() (err error) {
	p := f.page
	if p == nil {
		return
	}
	if p.dirty == false {
		return
	}
	if p.locked == true {
		panic(fmt.Errorf("Internal error: flushing locked page #%v\n", p.number))
	}
	_, err = f.pf.WritePage(p.buf, p.number)
	p.dirty = false
	if err != nil {
		return err
	}
//$$$	f.updateEntry()
return
}

func (f *DynamicPagedFile) freePage() (err error) {
	err = f.flushPage()
	if err != nil {
		return
	}
	f.page = nil
return
}

func (f *DynamicPagedFile) lockPage(num int64) (p *dynamicPage, err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.page != nil {
		if f.page.number == num {
			f.page.locked = true
			p = f.page
			return
		}
		if f.page.dirty == true {
			err = f.freePage()
			if err != nil {
				return nil, err
			}
		}
	}
	p = f.newPage()
	rd, err := f.pf.ReadPage(p.buf, num)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return nil, err
	}
//	p.parent = f
	p.number = num
	p.used	 = uint(rd)
	p.dirty	 = false
	p.locked = true
	f.page = p
return
}

func (f *DynamicPagedFile) unlockPage() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.page != nil {
		f.page.locked = false
	}
}

func (f *DynamicPagedFile) updateFileSize() { // For RW files
	p := f.page
	if p == nil {
		return
	}
	if p.number > f.lastpage {
		f.lastpage = p.number
		f.lastpagesize = p.used
	}
	if p.number == f.lastpage && p.used > f.lastpagesize {
		f.lastpagesize = p.used
	}
//$$$	f.updateEntry()
}

// dynamicPage

type dynamicPage struct {
//	parent		*DynamicPagedFile
	number		int64
	buf		[]byte
	used		uint
	dirty  		bool
	locked		bool
}

func (p *dynamicPage) readAt(b []byte, ofst uint) (rc int) {
	return copy(b, p.buf[ofst:p.used])
}

func (p *dynamicPage) writeAt(b []byte, ofst uint) (rc int) {
	rc = copy(p.buf[ofst:], b)
	if rc > 0 {
		pused := ofst + uint(rc)
		if pused > p.used {
			p.used = pused
		}
//$$$		p.updateFileSize() --> f.WriteAt
		p.dirty = true
	}
return
}

// Check interfaces
var (
	_ Node = &DynamicPagedFile{}
	_ File = &DynamicPagedFile{}
)
