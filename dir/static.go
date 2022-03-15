package dir

import (
	"time"
	"io/fs"
)

// Static Dir

type Static struct {
	name		string
	Child		[]fs.DirEntry
	mtime		time.Time
}

func NewStatic(name string, c []fs.DirEntry) *Static {
	return &Static {
		name:	name,
		Child:	c,
		mtime:	time.Now(),
	}
}

// fs.ReadDirFile
func (p *Static) ReadDir(n int) (res []fs.DirEntry, err error) {
return res, nil//$$$
}

// fs.File
func (p *Static) Stat() (fs.FileInfo, error) {
return p, nil
}

func (p *Static) Read(b []byte) (int, error) {
return 0, fs.ErrInvalid
}

func (p *Static) Close() error {
return nil
}

// fs.FileInfo
func (p *Static) Name() string {
return "."
}

func (p *Static) Size() int64 {
return int64(len(p.Child))
}

func (p *Static) Mode() fs.FileMode {
return fs.ModeDir | fs.ModePerm
}

func (p *Static) ModTime() time.Time {
return p.mtime
}

func (p *Static) IsDir() bool {
return true
}

func (p *Static) Sys() interface{} {
return p
}

// Check interfces
var _ fs.ReadDirFile = &Static{}
