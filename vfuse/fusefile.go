package vfuse

// Files

func (s *FS) Create(path string, flags int, mode uint32) (errc int, h uint64) { // 'flags' ignored!
	defer trace(LogCreate, path, flags, mode)(&errc, &h)
	defer s.sync()()
	h, errc = s.openCntPath(path, false)
	if errc == 0 { // Already Exists & Open
		//$$$ Truncate?
		return
	}
	n, errc := s.makeNode(path, mode)
	if errc != 0 {
		return errc, ^uint64(0)
	}
return 0, s.openCntNode(path, n)
}

func (s *FS) Open(path string, flags int) (errc int, h uint64) {
	defer trace(LogOpen, path, flags)(&errc, &h)
	defer s.sync()()
	h, errc = s.openCntPath(path, false)
	if errc != 0 {
		h = ^uint64(0)
	}
return
}

func (s *FS) Truncate(path string, size int64, h uint64) (errc int) {
	defer trace(LogTruncate, path, size, h)(&errc)
	defer s.sync()()
	h, errc = s.openCntPath(path, false)
	if errc != 0 {
		return
	}
	defer s.closeCntHandle(h)
	file, errc := s.getOpenFile(h)
	if errc != 0 { // That would be veeery strange!
		return
	}
return file.Truncate(size)
}

func (s *FS) Read(path string, b []byte, ofst int64, h uint64) (n int) {
	defer trace(LogRead, path, len(b), ofst, h)(&n)
	defer s.sync()()
	file, errc := s.getOpenFile(h)
	if errc != 0 {
		return
	}
	n, _ = file.ReadAt(b, ofst)
return
}

func (s *FS) Write(path string, b []byte, ofst int64, h uint64) (n int) {
	defer trace(LogWrite, path, len(b), ofst, h)(&n)
	defer s.sync()()
	file, errc := s.getOpenFile(h)
	if errc != 0 {
		return
	}
	n, _ = file.WriteAt(b, ofst)
return
}

// Flush flushes cached file data.
func (s *FS) Flush(path string, h uint64) (errc int) {
	defer trace(LogFlush, path, h)(&errc)
	defer s.sync()()
return s.fsync(path, false, h)
}

// Fsync synchronizes file contents.
func (s *FS) Fsync(path string, datasync bool, h uint64) (errc int) {
	defer trace(LogFsync, path, datasync, h)(&errc)
	defer s.sync()()
return s.fsync(path, datasync, h)
}

func (s *FS) fsync(path string, datasync bool, h uint64) (errc int) {
	node, errc := s.getOpenNode(h)
	if errc != 0 {
		return errc
	}
	node.DataSync()
	if datasync {
		return 0
	}
	node.Sync()
return 0
}

func (s *FS) Release(path string, h uint64) (errc int) {
	defer trace(LogRelease, path, h)(&errc)
	defer s.sync()()
return s.closeCntHandle(h)
}

func (s *FS) Unlink(path string) (errc int) {
	defer trace(LogUnlink, path)(&errc)
	defer s.sync()()
return s.removeNode(path)
}
