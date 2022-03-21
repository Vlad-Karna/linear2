package node

import (
	"time"

	. "github.com/Vlad-Karna/vfuse/vfuse"

	"github.com/billziss-gh/cgofuse/fuse"
)

type StaticBase struct {
}

func (p *StaticBase) Utime(t []time.Time) (errc int) {
return -fuse.EROFS
}

func (p *StaticBase) Remove() (errc int) {
return -fuse.EROFS
}

// StaticDir

type StaticDir struct {
	DirBase
	StaticBase
	Child		map[string]Node
}

func NewStaticDir(c map[string]Node) *StaticDir {
	return &StaticDir {
		Child: c,
	}
}

func (p *StaticDir) Lookup(n string) (res Node) {
	res, _ = p.Child[n]
return
}

func (p *StaticDir) Readdir(fill func(name string) bool) (errc int) {
	for n, _ := range p.Child {
		if !fill(n) {
			break
		}
	}
return 0
}

func (p *StaticDir) Mkdir(n string) (errc int) { //$$$ Phase out
return -fuse.EROFS
}

func (p *StaticDir) Make(n string, moude uint32) (res Node, errc int) {
return nil, -fuse.EROFS
}

func (p *StaticDir) Rename(node Node, newname string) (errc int) {
return -fuse.EROFS
}

// StaticFile

type StaticFile struct {
	FileBase
	StaticBase
	Data		[]byte
}

func NewStaticFile(b []byte) *StaticFile {
	return &StaticFile {
		Data: b,
	}
}

func (p *StaticFile) Getattr(stat *fuse.Stat_t) (errc int) {
	p.FileBase.Getattr(stat)
	stat.Size = int64(len(p.Data))
return 0
}

func (p *StaticFile) ReadAt(b []byte, off int64) (n int, err error) {
	if off >= int64(len(p.Data)) {
		return 0, nil
	}
	n = copy(b, p.Data[int(off):])
return
}

func (p *StaticFile) Truncate(sz int64) (errc int) {
return -fuse.EROFS
}

// Check interfaces
var (
	_ Node = &StaticDir{}
	_ Dir  = &StaticDir{}
	_ Node = &StaticFile{}
	_ File = &StaticFile{}
)
