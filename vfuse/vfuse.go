package vfuse

import (
	"log"
	"io"
	"sync"
	"time"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
)

type NodeType uint

const (
	DirNodeType		NodeType = 1 + iota
	FileNodeType
	LinkNodeType
)

// Abstract Classes

type Node interface {
	Type() NodeType
	Getattr(stat *fuse.Stat_t) (errc int)
	Listxattr(fill func(n string) bool) (errc int)
	Setxattr(name string, value []byte, flags int) (errc int)
	Getxattr(name string) (errc int, res []byte)
	Utime(t []time.Time) (errc int)
	Remove() (errc int)
	Sync()					// Sync Node Info
	DataSync()				// Sync all File Data / Dir's Children Node Info
}

type Dir interface {
	Node
	Lookup(n string) Node
	Readdir(fill func(name string) bool) (errc int)
	Make(n string, mode uint32) (res Node, errc int)
	Rename(node Node, newname string) (errc int)
	// Returns true if Dir's parent .Get()/.Put() should be called
//	Get() bool
//	Put() bool
}

type File interface {
	Node
	io.ReaderAt
	io.WriterAt
	io.Closer
	Truncate(sz int64) (errc int)
}

// FS

type FS struct {
	fuse.FileSystemBase
	mutex			sync.Mutex
	root			Dir
	handle			uint64
	OpenNode		map[uint64]*OpenNodeEntry	// maps handle to OpenNodeEntry struct
	openPath		map[string]uint64		// maps path to handle
	uid, gid		uint32
	mount			map[string]Dir			// mount points 'path' --> Dir
//	xmount			map[Dir]Dir
}

type OpenNodeEntry struct {
	Node
	handle		uint64
	path		string
	openCnt		uint
}

func NewFS(root Dir) (res *FS, err error) {
	res = &FS {
		root: root,
		handle: 1,
		OpenNode: make(map[uint64]*OpenNodeEntry),
		openPath: make(map[string]uint64),
		mount:    make(map[string]Dir),
	}
//	res.OpenNode[res.handle] = root;
//	res.openPath[""] = res.handle;
return
}

func (s *FS) sync () func() {
	s.mutex.Lock()
	return func() {
		s.mutex.Unlock()
	}
}

func (s *FS) lookup(path string) (res Node, rpath []string, errc int) {
	res, rpath, _, errc = s.lookupMount(nil, path)
return
}

func (s *FS) lookupExt(path string, proh Node) (res Node, rpath []string, pdir Dir, errc int) { // Lookup 'path' Node, watch for the prohibited 'proh' Node on the way
return s.lookupMount(proh, path)
}

func (s *FS) lookupMount(proh Node, path string) (res Node, rpath []string, pdir Dir, errc int) {
	from := s.root
// Check Mount Points
	for p, d := range s.mount {
		if strings.HasPrefix(path, p) {
			from = d
			path = path[len(p):]
		}
	}
// Lookup From Root
return s.lookupFrom(from, proh, strings.Split(path, "/")...)
}

func (s *FS) lookupFrom(curnode Node, proh Node, path... string) (res Node, rpath []string, pdir Dir, errc int) {
	for len(path) > 0 && errc == 0 {
		if path[0] == "" {
			path = path[1:]
			continue
		}
		if curnode == proh { // Prohibited Node found. Circular dependency!
			errc = -fuse.ELOOP
			break
		}
		if curnode.Type() != DirNodeType { // Not Dir
			errc = -fuse.ENOTDIR
			break
		}
		dir := curnode.(Dir)
//$$$ dir.open -- cnt++
		c := dir.Lookup(path[0])
//$$$ dir.close -- cnt--
		if c == nil {
			errc = -fuse.ENOENT
			break
		}
		curnode, pdir, path = c, dir, path[1:]
	}
return curnode, path, pdir, errc
}

func (s *FS) getNode(p string) (n Node, on *OpenNodeEntry, errc int) {
	h, ok := s.openPath[p]
	if ok { // Already Open
		on = s.OpenNode[h]
		n = on.Node
	} else { // Not Yet Open
		n, _, errc = s.lookup(p)
	}
	trace(LogAll, p)(n, on, errc)//$$$
return
}

func (s *FS) openCntNode(path string, n Node) (h uint64) {
	on := &OpenNodeEntry { Node: n, handle: s.handle, path: path, openCnt: 1 }
	s.OpenNode[s.handle] = on
	s.openPath[path] = s.handle
	h = s.handle
	if  LogMask & LogOpenAny != 0 {
		log.Printf("openCntNode: %+v\n", on)
	}
	s.handle++
return
}

func (s *FS) openCntPath(p string, isdir bool) (h uint64, errc int) {
	defer trace(LogOpenAny, p, isdir)(&h, &errc)
	n, on, errc := s.getNode(p)
	if errc != 0 {
		return
	}
	if on != nil { // Already Open
		h = on.handle
		on.openCnt++
	} else { // Not Yet Open
		switch {
		case n.Type() == DirNodeType && !isdir: return 0, -fuse.EISDIR
		case n.Type() != DirNodeType &&  isdir: return 0, -fuse.ENOTDIR
		}
		h = s.openCntNode(p, n)
	}
return
}

func (s *FS) closeCntHandle(h uint64) (errc int) {
	defer trace(LogReleaseAny, h)(&errc)
	on, ok := s.OpenNode[h]
	if !ok {
		return -fuse.EINVAL
	}
	if on.openCnt > 0 {
		on.openCnt--
	}
	if on.openCnt == 0 {
		delete(s.openPath, on.path)
		delete(s.OpenNode, h)
		switch n := on.Node.(type) {
		case File:
			n.Close()
		default:
			on.Node.Sync()
		}
		if LogMask & LogReleaseAny != 0 {
			log.Printf("closeCntHandle(%v): %v closed.\n", h, on.path)
		}
	}
return 0
}

func (s *FS) getOpenNode(h uint64) (res Node, errc int) {
	on, ok := s.OpenNode[h]
	if !ok {
		errc = -fuse.EINVAL
	}
return on.Node, 0
}

func (s *FS) getOpenDir(h uint64) (res Dir, errc int) {
	n, errc := s.getOpenNode(h)
	if errc != 0 {
		return
	}
	res, ok := n.(Dir)
	if !ok {
		errc = -fuse.EINVAL
	}
return
}

func (s *FS) getOpenFile(h uint64) (res File, errc int) {
	n, errc := s.getOpenNode(h)
	if errc != 0 {
		return
	}
	res, ok := n.(File)
	if !ok {
		errc = -fuse.EINVAL
	}
return
}

// Makes specified Node (File or Dir) and returns it
// Creates and opens FileNode in the Host Local Storage if it's a File
func (s *FS) makeNode(path string, mode uint32) (res Node, errc int) {
	defer trace(LogMake, path, mode)(&errc)
	lg := false
	if LogMask & LogMake != 0 {
		lg = true
		log.Printf("makeNode   %v\n", path)
	}
	n, p, errc := s.lookup(path)
	if errc == 0 {
		if lg {
			log.Printf("makeNode:   %v: EXISTS\n", path)
		}
		return n, -fuse.EEXIST
	}
	if len(p) > 1 {
		if lg {
			log.Printf("makeNode:   %v: %v (p:%v)\n", path, errc, p)
		}
		return
	}
return n.(Dir).Make(p[0], mode)
}

// Removes specified Node (File or Dir) if possible
func (s *FS) removeNode(path string) (errc int) {
	n, _, errc := s.lookup(path)
	if errc != 0 {
		if LogMask & LogRemoveAny != 0 {
			log.Printf("removeNode  %v: %v\n", path, errc)
		}
		return errc
	}
return n.Remove()
}
