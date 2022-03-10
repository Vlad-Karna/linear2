package linear

import (
	"time"
	"io/fs"
)

type Linear struct {
	file			fs.File
	bs1, bs2, bs, blocks	int64
}

func NewLinear(f fs.File, bs1, bs2 int64) (res fs.File, err error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	sz := st.Size()

	return &Linear {
		file:	f,
		bs1:	bs1,
		bs2:	bs2,
		bs:	bs1 + bs2,
		blocks:	sz/(bs1 + bs2),
	}, nil
}

// FileInfo
func (f *Linear) Name() string {
	st, err := f.Stat()
	if err != nil {
		return ""
	}
return st.Name()
}

func (f *Linear) Size() int64 {
return f.bs * f.blocks
}

func (f *Linear) Mode() fs.FileMode {
	st, err := f.Stat()
	if err != nil {
		return 0
	}
return st.Mode()
}

func (f *Linear) ModTime() (res time.Time) {
	st, err := f.Stat()
	if err != nil {
		return
	}
return st.ModTime()
}

func (f *Linear) IsDir() bool {
return false
}

func (f *Linear) Sys() interface{} {
return f
}

// File
func (f *Linear) Stat() (res fs.FileInfo, err error) {
return f, nil
}

func (f *Linear) Read(p []byte) (n int, err error) {
return f.ReadAt(p, 0)
}

func (f *Linear) Close() error {
return f.file.Close()
}

// ReaderAt
func (f *Linear) ReadAt(p []byte, off int64) (n int, err error) {
	for n = 0; cap(p) > 0; {
		bn, bo := off / f.bs, off % f.bs	// block number, block offset
		o, tr := bn * f.bs + bo, f.bs - bo	// Read no more than 1 block
		var nn int
		nn, err = f.ReadAt(p[:tr], o)
		if err != nil {
			return
		}
		n   += nn
		off += int64(nn)
		p = p[nn:]
	}
return
}
