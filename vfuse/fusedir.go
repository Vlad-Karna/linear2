package vfuse

import (
	"path/filepath"

	"github.com/billziss-gh/cgofuse/fuse"
)

// Dirs

func (s *FS) Opendir(path string) (errc int, h uint64) {
	defer trace(LogOpendir, path)(&errc, &h)
	defer s.sync()()
	h, errc = s.openCntPath(path, true)
	if errc != 0 {
		h = ^uint64(0)
	}
return
}

func (s *FS) Readdir (path string, fill func(name string, stat *fuse.Stat_t, ofst int64) bool, ofst int64, h uint64) (errc int) {
	defer trace(LogReaddir, path, ofst, h)(&errc)
	defer s.sync()()
	dir, errc := s.getOpenDir(h)
	if errc != 0 {
		return
	}
	fill (".",  nil, 0)
	fill ("..", nil, 0)
return dir.Readdir(func (n string) bool{
	return fill (n, nil, 0)
})
}

// Fsyncdir synchronizes directory contents
func (s *FS) Fsyncdir(path string, datasync bool, h uint64) (errc int) {
	defer trace(LogFsyncdir, path, datasync, h)(&errc)
	defer s.sync()()
return s.fsync(path, datasync, h)
}

func (s *FS) Releasedir(path string, h uint64) (errc int) {
	defer trace(LogReleasedir, path, h)(&errc)
	defer s.sync()()
return s.closeCntHandle(h)
}

func (s *FS) Mkdir(path string, mode uint32) (errc int) {
	defer trace(LogMkdir, path, mode)(&errc)
	defer s.sync()()
	_, errc = s.makeNode(path, mode | fuse.S_IFDIR)
return
}

func (s *FS) Rmdir(path string) (errc int) {
	defer trace(LogRmdir, path)(&errc)
	defer s.sync()()
return s.removeNode(path)
}

func (s *FS) Rename(oldpath string, newpath string) (errc int) {
	defer trace(LogRename, oldpath, newpath)(&errc)
	defer s.sync()()
	oldnode, _, errc := s.lookup(oldpath)
	if errc != 0 {
		return
	}
	newnode, newrpath, newpdir, errc := s.lookupExt(newpath, oldnode)
	if errc == -fuse.ENOENT && len(newrpath) == 1 { // 'newpath' not exists, but new dir does. it's ok.
		errc = 0
		newpdir, newnode = newnode.(Dir), nil
	}
	if errc != 0 {
		return
	}
	if newnode != nil { // Delete Target Node If Exists
		errc = newnode.Remove()
		if errc != 0 {
			return
		}
	}
return newpdir.Rename(oldnode, filepath.Base(newpath))
}
