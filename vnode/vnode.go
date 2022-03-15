package vnode

import (
	"os"
	"log"
	"errors"
	"strings"
	"io/fs"

	"github.com/billziss-gh/cgofuse/fuse"

	"../vfuse"
	"../dir"
	"../file"
)

// Base: Base class for vfuse.Dir/File
// NOT a Node implementation
type Base struct {
	fs.File
}

func (p Base) Type() vfuse.NodeType {
	if p.File.IsDir() {
		return vfuse.DirNodeType
	} else {
		return vfuse.FileNodeType
	}
}

func (p Base) Getattr(stat *fuse.Stat_t) (errc int) {
	*stat = fuse.Stat_t {
		Mode:	mapMode(p.Mode()),
		Nlink:	1,
		Size:	p.Size,
		Blksize:vfuse.FsBlockSize,
		Atim:	fuse.NewTimespec(p.Atime),
		Mtim:	fuse.NewTimespec(p.Mtime),
		Ctim:	fuse.NewTimespec(p.Ctime),
		Birthtim: fuse.NewTimespec(p.Btime),
		// Uid, Gid will be set by vfuse.Getattr
	}
return 0
}

func mapMode(osm os.FileMode) (res uint32) {
	res = uint32(osm & os.ModePerm)
	switch osm & os.ModeType {
	case 0:			res |= fuse.S_IFREG
	case os.ModeDir:	res |= fuse.S_IFDIR
	case os.ModeSymlink:	res |= fuse.S_IFLNK
	}
return
}

func mapError(err error) (errc int) { // Unused by Base & Dir
	switch {
	case err == nil: return 0
	case errors.Is(err, os.ErrNotExist):	return -fuse.ENOENT
//	case errors.Is(err, basedir.ErrNotDir):	return -fuse.ENOTDIR
//	case errors.Is(err, file.ErrMac):	return -fuse.EIO
	}
log.Printf("mapError   %v --> EIO\n", err)
return -fuse.EIO
}

func (p Base) Listxattr(fill func(n string) bool) (errc int) {
	if !fill(XattrID) { // ID
		return -fuse.ERANGE
	}
	if !fill(XattrNode) { // Node
		return -fuse.ERANGE
	}
	for pkey := range p.ACL { // ACL
		if !fill(XattrPKeyPfx + pkey.String()) {
			return -fuse.ERANGE
		}
	}
	//$$$ Genneral Purpose Xattrs
return 0
}

func (p Base) Setxattr(name string, value []byte, flags int) (errc int) {
	if p.RO {
		return -fuse.EACCES
	}
	if !strings.HasPrefix(name, Xattr) { // General Purpose Xattrs
		//$$$ Genneral Purpose Xattrs
		return -fuse.ENOATTR
	}
// S4 Specific Xattrs
return -fuse.EPERM
}

func (p Base) Getxattr(name string) (errc int, res []byte) {
	if !strings.HasPrefix(name, Xattr) { // General Purpose Xattrs
		//$$$ Genneral Purpose Xattrs
		return -fuse.ENOATTR, nil
	}
// S4 Specific Xattrs
	var err error
	switch {
	case name == XattrID: // ID
		res = []byte(p.ID.String())
	case name == XattrNode: // Node
		nn := *p.NodeInfo
		nn.Parent, nn.Content = nil, nil
		if res, err = json.Marshal(&nn); err != nil {
			errc = -fuse.ENOATTR
		}
	case strings.HasPrefix(name, XattrPKeyPfx): // ACL/PKey
		var pk keys.PKey
		err = pk.UnmarshalText([]byte(name[len(XattrPKeyPfx):]))
		errc = -fuse.ENOATTR // Will set to 0 later, if everything's ok
		if err == nil {
			cont, ok := p.ACL[pk]
			if ok {
				res = []byte(cont)
				errc = 0
			}
		}
	}
return
}

func (p Base) Remove() (errr int) {
	if p.RO {
		return -fuse.EACCES
	}
// Delete From Parent Dir
	pdir := p.ParentNode().(*node.DirNode) // Couldn't be BaseDir. Root at best.
	err := pdir.DelChild(p.NodeInfo)
	if err != nil {
		return -fuse.ENOENT // The only err returned by pdir.DelChild
	}
/*// Delete Data (or XX/ Dir if needed)
	err = pdir.RemoveNodeData(p.NodeInfo)
	if err != nil {
		log.Printf("Base.Remove  : RemoveHostData: %v\n", p)
		// Ignoring error: if DirNodeType & Hierarchical Name Mapping (not empty dirs w/ *.node files)
		err = nil
	}*/
	pdir.RemoveNodeInfo(p.NodeInfo) //leaving Dtime to sync.ProxyStorage
return
}

func (p Base) Sync() {
	p.Update()
}

// Dir

type Dir struct { // vfuse Dir implementation
	Base
}

func (p *Dir) Getattr(stat *fuse.Stat_t) (errc int) {
	errc = p.Base.Getattr(stat)
	if errc != 0 {
		return
	}
	stat.Size = vfuse.FsBlockSize
return
}

func (p *Dir) load() (res *node.DirNode) {
	if p.Content != nil {
		res = p.Content.(*node.DirNode)
	} else {
		res = node.NewDirNode(p.NodeInfo)
		res.Read()
		p.Content = res
	}
	res.Read() // Read/Refresh-if-needed
return
}

func (p *Dir) Lookup(n string) vfuse.Node {
	d := p.load()
	res, err := d.Lookup(n)
	if err != nil {
		return nil
	}
return mknode(res)
}

func mknode(ni *node.NodeInfo) (res vfuse.Node) {
	b := Base { ni }
	switch ni.Type {
	case node.DirNodeType:  return &Dir  { b }
	case node.FileNodeType: return &File { b }
	}
return nil
}

func (p *Dir) Readdir(fill func(name string) bool) (errc int) {
	d := p.load()
	for n, _ := range d.Child {
		if !fill(n) {
			break
		}
	}
return 0
}

func (p *Dir) Make(n string, mode uint32) (res vfuse.Node, errc int) {
	if p.RO {
		return nil, -fuse.EACCES
	}
	d := p.load()
	if res, ok := d.Child[n]; ok {
		return mknode(res), -fuse.EEXIST
	}
	var typ node.NodeType
	if mode & fuse.S_IFDIR != 0 {
		typ = node.DirNodeType
	} else {
		typ = node.FileNodeType
	}
//$$$ Link
	ni, err := d.NewNodeInfo(d, typ, n) // New NodeInfo w/ Initialized PAC/ACL
	if err != nil {
		return nil, mapError(err)
	}

// Perform Changes on Local Storage
	switch typ {
	case node.DirNodeType:
		ni.Content = node.NewDirNode(ni)
	case node.FileNodeType:
		fd, err := d.CreateNodeData(ni)
		if err != nil {
			return nil, mapError(err)
		}
		ni.Content, err = node.NewFileNode(ni, fd) // Should be OK for the newly created file
		if err != nil {
			return nil, mapError(err)
		}
	}

// Change Parent Dir
	err = d.AddChild(ni)
	if err != nil { // The only error AddChild throws is EEXIST
		log.Printf("ERROR: Dir.Make(%v, %v): AddChild: %v\n", n, mode, err)
		return nil, mapError(err)
	}
	err = d.SaveNodeInfo(ni)
	if err != nil { // Shouldn't be any error at all
		log.Printf("ERROR: Dir.Make(%v, %v): SaveNodeInfo: %v\n", n, mode, err)
		d.DelChild(ni)
		return nil, mapError(err)
	}
return mknode(ni), 0
}

func (p *Dir) Remove() (errc int) {
	if p.RO {
		return -fuse.EACCES
	}
	d := p.load()
	if len(d.Child) > 0 {
		return -fuse.ENOTEMPTY
	}
return p.Base.Remove()
}

// Rename 'node' to 'name'.
// Re-parent 'node' from whatever Directory it is to 'p' if needed. And
// Update NodeInfo if 'node' is p's Child.
func (p *Dir) Rename(nod vfuse.Node, name string) (errc int) {
	if p.RO {
		return -fuse.EACCES
	}
	var ni *node.NodeInfo
	switch n := nod.(type) {
	case *File: ni = n.NodeInfo
	case *Dir:  ni = n.NodeInfo
	//$$$ *Link
	}
	d := p.load()
	_, local := d.ChildById[ni.ID]
	if local {
		log.Printf("Renaming   '%s' --> '%s'\n", ni.Name, name)
	} else {
		log.Printf("Moving     '%s' --> '%s'\n", ni.Name, name)
	}
	copy := *ni
// Change NodeInfo
	nichanged := false
	if ni.Name != name {
		ni.Name = name
		nichanged =  true
	}
	if local { // Same parent
		d.DelChild(&copy)
		d.AddChild(ni)
	} else { // New parent
		ni.Parent = d
		err := d.AddChild(ni)
		if err != nil { // The only error AddChild throws is EEXIST
			log.Printf("ERROR: Dir.Rename(%v, %v): AddChild: %v\n", ni.ID, name, err)
			return mapError(err)
		}
		ni.PAC = nil // Will Be Created w/ New Parent at ni.Save
	}
// Save New NodeInfo. New Host Local Storage Dirs are created.
	if nichanged {
		ni.Change()
	}
	d.SaveNodeInfo(ni)

	if !local {
		err := d.MoveNodeData(&copy, ni) // Same NodeID ==> for *.data, same FileID ==> Same File Name
		if err != nil {
			log.Printf("ERROR: Dir.Rename(%v, %v): MoveData: %v\n", ni.ID, name, err)
			return mapError(err)
		}
		// Remove from Old Parent
		oldpdir := copy.ParentNode().(*node.DirNode)
		oldpdir.DelChild(&copy)
		// Remove Old NodeInfo
		err = d.RemoveNodeInfo(&copy)
		if err != nil { // Shouldn't be any
			log.Printf("ERROR: Dir.Rename(%v, %v): RemoveNodeInfo: %v\n", ni.ID, name, err)
			return mapError(err)
		}
	}
return 0
}

func (p *Dir) DataSync() {
	d := p.load()
	for _, ni := range d.Child {
		ni.Update()
	}
}

// File

type File struct { // vfuse File implementation
	Base
}

func (p *File) load() (res *node.FileNode, err error) {
	if p.Content != nil {
		res = p.Content.(*node.FileNode)
	} else {
		fd, err := p.Parent.OpenNodeData(p.NodeInfo)
		if err != nil {
			return nil, err
		}
		res, err = node.NewFileNode(p.NodeInfo, fd)
		if err != nil {
			return nil, err
		}
		p.Content = res
	}
return
}

func (p *File) ReadAt(b []byte, off int64) (n int, err error) {
	f, err := p.load()
	if err != nil {
		return
	}
return f.ReadAt(b, off)
}

func (p *File) WriteAt(b []byte, off int64) (n int, err error) {
	if p.RO {
		return 0, os.ErrPermission
	}
	f, err := p.load()
	if err != nil {
		return
	}
return f.WriteAt(b, off)
}

func (p *File) Close() error {
	if p.Content != nil {
		f := p.Content.(*node.FileNode)
		f.Close()
		p.Cleanup()
	}
return nil
}

func (p *File) Truncate(sz int64) (errc int) {
	if p.RO {
		return -fuse.EACCES
	}
	f, err := p.load()
	if err != nil {
		return
	}
	err = f.Truncate(sz)
return mapError(err)
}

func (p *File) Remove() (errr int) {
	if p.RO {
		return -fuse.EACCES
	}
// Delete Data (or XX/ Dir if needed)
	if err := p.ParentNode().(*node.DirNode).RemoveNodeData(p.NodeInfo); err != nil {
		log.Printf("Base.Remove  : RemoveHostData: %v\n", p)
		// Ignoring error: if DirNodeType & Hierarchical Name Mapping (not empty dirs w/ *.node files)
	}
return p.Base.Remove()
}

func (p *File) DataSync() {
	file := p.Content
	if file != nil {
		file.(*node.FileNode).Sync()
	}
}

// Check interfaces
var (
	_ vfuse.Node = &Dir{}
	_ vfuse.Dir  = &Dir{}
	_ vfuse.Node = &File{}
	_ vfuse.File = &File{}
)
