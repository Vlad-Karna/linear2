package file

import (
	"time"
	"io"
	"io/fs"
)

// Static File

type Static struct {
	Data		[]byte
	name		string
	mtime		time.Time
}

func NewStatic(b []byte, name string) *Static {
	return &Static {
		Data:	b,
		name:	name,
		mtime:	time.Now(),
	}
}

// fs.File
func (p *Static) Stat() (fs.FileInfo, error) {
return p, nil
}

func (p *Static) Read(b []byte) (int, error) {
	if (len(b) == len(p.Data)) {
		copy(b, p.Data)
		return len(b), io.EOF
	}
return 0, io.EOF
}

func (p *Static) Close() error {
return nil
}

// fs.FileInfo
func (p *Static) Name() string {
return p.name
}

func (p *Static) Size() int64 {
return int64(len(p.Data))
}

func (p *Static) Mode() fs.FileMode {
return fs.ModePerm
}

func (p *Static) ModTime() time.Time {
return p.mtime
}

func (p *Static) IsDir() bool {
return false
}

func (p *Static) Sys() interface{} {
return p
}

//io.ReaderAt
func (p *Static) ReadAt(b []byte, off int64) (n int, err error) {
	if off >= int64(len(p.Data)) {
		return 0, nil
	}
	n = copy(b, p.Data[int(off):])
return
}

// Check interfaces
var (
	_ fs.File	= &Static{}
	_ io.ReaderAt	= &Static{}
)
