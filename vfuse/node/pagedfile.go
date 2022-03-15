package node

import (
	"io"
	"io/fs"
	"errors"

	. "github.com/Vlad-Karna/vfs/vfuse"
)

var ErrNotReaderWriterAt = errors.New("neither io.ReaderAt nor io.WriterAt")

type PagedFile struct {
	fs.File		// w/ io.ReaderAt, io.WriterAt
	offset, size	int64
	pageSize	int
	pages		int64
}

func NewPagedFile(f fs.File, offset, size int64, pgsz int) (res *PagedFile, err error) {
	if f.(io.ReaderAt) == nil || f.(io.WriterAt) == nil {
		return nil, ErrNotReaderWriterAt
	}
//$$$ Check f.Size(), offset, size
// adjust
	res = &PagedFile{ File: f, offset: offset, size: size, pageSize: pgsz, pages: size/int64(pgsz)}
return
}

func (p *PagedFile) PageSize() int {
return p.pageSize
}

func (p *PagedFile) Size() int64 {
return p.size
}

func (p *PagedFile) ReadPage (b []byte, page int64) (n int, err error) {
	if page > p.pages {
		return 0, fs.ErrInvalid
	}
	if len(b) > p.pageSize {
		b = b[:p.pageSize]
	}
return p.File.(io.ReaderAt).ReadAt(b, p.offset + page * int64(p.pageSize))
}

func (p *PagedFile) WritePage(b []byte, page int64) (n int, err error) {
	if page > p.pages {
		return 0, fs.ErrInvalid
	}
	if len(b) > p.pageSize {
		b = b[:p.pageSize]
	}
return p.File.(io.WriterAt).WriteAt(b, p.offset + page * int64(p.pageSize))
}

// Check interfaces
var (
	_ PagedFileInterface = &PagedFile{}
)
