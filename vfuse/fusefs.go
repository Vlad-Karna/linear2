package vfuse

import (
	"os"

	"github.com/billziss-gh/cgofuse/fuse"
)

const FsBlockSize = 0x1000

// Init

func (s *FS) Init() {
	s.uid = uint32(os.Geteuid())
	s.gid = uint32(os.Getegid())
	trace(LogAll, "Uid: ", s.uid, "Gid: ", s.gid)()
	s.openCntPath("/", true)
}

func (s *FS) Destroy() {
	s.closeCntHandle(1) // Root Dir's Handle is Always '1'
	s.root = nil
}

func (s *FS) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	defer trace(LogStatfs, path)(stat, &errc)
	defer s.sync()()
	*stat = fuse.Statfs_t{
		Bsize:	FsBlockSize,
		Frsize:	FsBlockSize,
		Blocks:	0x100000000000,
		Bfree:	0x100000000000,
		Bavail:	0x100000000000,
		Files:	0x100000000000,
		Ffree:	0x100000000000,
		Favail:	0x100000000000,
		Namemax:	255,
	}
return 0
}

func (s *FS) Mount(path string, dir Dir) (errc int) {
	defer trace(LogMount, path, dir)(&errc)
	defer s.sync()()
	_, ok := s.mount[path]
	if ok {
		return -fuse.EBUSY
	}
	s.mount[path] = dir
return
}

func (s *FS) Unmount(path string) (errc int) {
	defer trace(LogUnmount, path)(&errc)
	_, ok := s.mount[path]
	if !ok {
		return -fuse.EINVAL
	}
	// Is 'd' (or it's sub-node) open?
	delete(s.mount, path)
return
}
