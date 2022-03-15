package vfuse

import (
	"strings"
	"flag"

	logtrace "github.com/Vlad-Karna/vfs/trace"
)

//go:generate enumer -type LogMaskType -trimprefix=Log
type LogMaskType uint64

const (
	LogGetattr	LogMaskType = 1 << iota
	LogChmod
	LogChown
	LogUtimens
	LogAccess

	LogOpendir
	LogReaddir
	LogFsyncdir
	LogReleasedir
	LogMkdir
	LogRmdir

	LogOpen
	LogCreate
	LogTruncate
	LogRead
	LogWrite
	LogRelease
	LogUnlink
	LogFlush
	LogFsync

	LogListxattr
	LogSetxattr
	LogGetxattr
	LogRemovexattr

	LogInit
	LogDestroy
	LogStatfs
	LogMake /*(!)*/
	LogRemove
	LogRename
	LogMount
	LogUnmount

	logLast

	LogEverything	LogMaskType = logLast - 1

	LogAttr		LogMaskType = LogGetattr | LogChmod | LogChown | LogUtimens | LogAccess
	LogXattr	LogMaskType = LogListxattr | LogSetxattr | LogGetxattr | LogRemovexattr
	LogDir		LogMaskType = LogOpendir | LogReaddir | LogFsyncdir | LogReleasedir | LogMkdir | LogRmdir
	LogFile		LogMaskType = LogOpen | LogCreate | LogTruncate | LogRead | LogWrite | LogRelease | LogUnlink | LogFlush | LogFsync
	LogFs		LogMaskType = LogInit | LogDestroy | LogStatfs | LogMake | LogRemove | LogRename | LogMount | LogUnmount

	LogAll		LogMaskType = LogAttr | LogDir | LogFile | LogXattr | LogFs

	LogOpenAny	LogMaskType = LogOpendir | LogOpen
	LogReleaseAny	LogMaskType = LogReleasedir | LogRelease
	LogCreateAny	LogMaskType = LogCreate | LogMkdir | LogMake /*(!)*/
	LogRemoveAny	LogMaskType = LogRmdir | LogUnlink | LogRemove
)

var LogMask LogMaskType = LogAll

type LogMaskSet LogMaskType

func (p *LogMaskSet) String() (res string) {
	v := LogMaskType(*p)
	res = ""
	for _, m := range LogMaskTypeValues() {
		if v & m != 0 {
			if len(res) > 0 {
				res += "+"
			}
			res += m.String()
		}
	}
return
}

func (p *LogMaskSet) Set(set string) error {
	v := LogMaskType(0)
	if set != "" {
		for _, s := range strings.Split(set, "+") {
			m, err := LogMaskTypeString(s)
			if err != nil {
				return err
			}
			v |= m
		}
	}
	*p = LogMaskSet(v)
return nil
}

// Check interface
var (
	setAll LogMaskSet = LogMaskSet(LogMask)
	_ flag.Value = &setAll
)

func trace(caller LogMaskType, vals ...interface{}) func(vals ...interface{}) {
	if LogMask & caller != 0 {
		return logtrace.Trace(1, vals...)
	}
return func(vals ...interface{}){}
}
