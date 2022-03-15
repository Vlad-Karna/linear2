package vfuse

// Xattrs

func (s *FS) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(LogListxattr, path)(&errc)
	defer s.sync()()
	n, _, errc := s.getNode(path)
	if errc != 0 {
		return
	}
return n.Listxattr(fill)
}

func (s *FS) Setxattr(path, name string, value []byte, flags int) (errc int) {
	defer trace(LogSetxattr, path, name, value, flags)(&errc)
	defer s.sync()()
	n, _, errc := s.getNode(path)
	if errc != 0 {
		return
	}
return n.Setxattr(name, value, flags)
}

func (s *FS) Getxattr(path, name string) (errc int, res []byte) {
	defer trace(LogGetxattr, path, name)(&errc, &res)
	defer s.sync()()
	n, _, errc := s.getNode(path)
	if errc != 0 {
		return
	}
return n.Getxattr(name)
}

func (s *FS) Removexattr(path string, name string) (errc int) {
return
}
