package node

import (
	"time"

	. "github.com/Vlad-Karna/vfs/vfuse"

	"github.com/billziss-gh/cgofuse/fuse"
)

// Base Classes

type NodeBase struct {
}

func (p *NodeBase) Getattr(stat *fuse.Stat_t) (errc int) {
	*stat = fuse.Stat_t {
		Nlink:	1,
		Blksize:FsBlockSize,
		Atim:	fuse.Now(),
		Mtim:	fuse.Now(),
		Ctim:	fuse.Now(),
		Birthtim: fuse.Now(),
	}
return 0
}

func (p *NodeBase) Listxattr(fill func(n string) bool) (errc int) {
return 0
}

func (p *NodeBase) Setxattr(name string, value []byte, flags int) (errc int) {
return 0
}

func (p *NodeBase) Getxattr(name string) (errc int, res []byte) {
return -fuse.ENOATTR, nil
}

func (p *NodeBase) Utime(t []time.Time) (errc int) {
//%%$$$
return 0
}

func (p *NodeBase) Remove() (errc int) {
return -fuse.ENOSYS
}

func (p *NodeBase) Sync() {
}

func (p *NodeBase) DataSync() {
}

type DirBase struct {
	NodeBase
}

func (d *DirBase) Type() NodeType {
return DirNodeType
}

func (p *DirBase) Getattr(stat *fuse.Stat_t) (errc int) {
	p.NodeBase.Getattr(stat)
	stat.Mode = fuse.S_IFDIR | 0755
return 0
}

func (p *DirBase) Lookup(n string) (res Node) {
return nil
}

func (p *DirBase) Readdir(fill func(name string) bool) (errc int) {
return 0
}

func (p *DirBase) Mkdir(n string) (errc int) {//$$$ Phase out
return -fuse.ENOSYS
}

func (p *DirBase) Make(n string, mode uint32) (res Node, errc int) {
return nil, -fuse.ENOSYS
}

func (p *DirBase) Rename(node Node, newname string) (errc int) {
return -fuse.ENOSYS
}

// FileBase

type FileBase struct {
	NodeBase
}

func (p *FileBase) Type() NodeType {
return FileNodeType
}

func (p *FileBase) Getattr(stat *fuse.Stat_t) (errc int) {
	p.NodeBase.Getattr(stat)
	stat.Mode = fuse.S_IFREG | 0644
return 0
}

func (p *FileBase) ReadAt(b []byte, off int64) (n int, err error) {
return 0, nil
}

func (p *FileBase) WriteAt(b []byte, off int64) (n int, err error) {
return len(b), nil
}

func (p *FileBase) Close() error {
return nil
}

func (p *FileBase) Truncate(sz int64) (errc int) {
return -fuse.ENOSYS
}

// Check interfaces
var (
	_ Node = &DirBase{}
	_ Dir  = &DirBase{}
	_ Node = &FileBase{}
	_ File = &FileBase{}
)
