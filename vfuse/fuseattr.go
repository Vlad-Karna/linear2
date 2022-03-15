package vfuse

import (
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
)

// Attrs

func (s *FS) Getattr(path string, stat *fuse.Stat_t, h uint64) (errc int) {
	defer trace(LogGetattr, path, h)(stat, &errc)
	defer s.sync()()
return s.getattr(path, stat, h)
}

func (s *FS) getattr(path string, stat *fuse.Stat_t, h uint64) (errc int) {
	node, _, errc := s.getNode(path)
	if errc != 0 {
		return
	}
	errc = node.Getattr(stat)
	if errc != 0 {
		return
	}
	stat.Uid = s.uid
	stat.Gid = s.gid
return
}

func (s *FS) Chmod(path string, mode uint32) (errc int) {
	defer trace(LogChmod, path, mode)(&errc)
	defer s.sync()()
//$$$ if mode == NO READ for GRP & OTHERS ==> clear ACL
return 0
}

func (s *FS) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(LogChown, path, uid, gid)(&errc)
	defer s.sync()()
//$$$ Have some UID -- PKey Mapping ==> change file owner (Me) to someone other
return 0
}

func (s *FS) Utimens(path string, times []fuse.Timespec) (errc int) {
	defer trace(LogUtimens, path, times)(&errc)
	defer s.sync()()
	n, _, errc := s.getNode(path)
	if errc != 0 {
		return
	}
return n.Utime([]time.Time {times[0].Time(), times[1].Time()})
}

func (s *FS) Access(path string, mask uint32) (errc int) {
	defer trace(LogAccess, path, mask)(&errc)
	defer s.sync()()
	var stat fuse.Stat_t
	errc = s.getattr(path, &stat, ^uint64(0))
	if errc != 0 {
		return
	}
	if (stat.Mode >> 6) & mask != mask {
		return -fuse.EACCES
	}
return 0
}
