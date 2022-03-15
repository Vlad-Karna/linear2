package trace

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

func traceJoin(deref bool, vals []interface{}) string {
	rslt := ""
	for _, v := range vals {
		if deref {
			switch i := v.(type) {
			case *bool:
				rslt += fmt.Sprintf(", %#v", *i)
			case *int:
				rslt += fmt.Sprintf(", %#v", *i)
			case *int8:
				rslt += fmt.Sprintf(", %#v", *i)
			case *int16:
				rslt += fmt.Sprintf(", %#v", *i)
			case *int32:
				rslt += fmt.Sprintf(", %#v", *i)
			case *int64:
				rslt += fmt.Sprintf(", %#v", *i)
			case *uint:
				rslt += fmt.Sprintf(", %#v", *i)
			case *uint8:
				rslt += fmt.Sprintf(", %#v", *i)
			case *uint16:
				rslt += fmt.Sprintf(", %#v", *i)
			case *uint32:
				rslt += fmt.Sprintf(", %#v", *i)
			case *uint64:
				rslt += fmt.Sprintf(", %#v", *i)
			case *uintptr:
				rslt += fmt.Sprintf(", %#v", *i)
			case *float32:
				rslt += fmt.Sprintf(", %#v", *i)
			case *float64:
				rslt += fmt.Sprintf(", %#v", *i)
			case *complex64:
				rslt += fmt.Sprintf(", %#v", *i)
			case *complex128:
				rslt += fmt.Sprintf(", %#v", *i)
			case *string:
				rslt += fmt.Sprintf(", %#v", *i)
			default:
				rslt += fmt.Sprintf(", %#v", v)
			}
		} else {
			rslt += fmt.Sprintf(", %#v", v)
		}
	}
	if len(rslt) > 0 {
		rslt = rslt[2:]
	}
	return rslt
}

func Trace(skip int, vals ...interface{}) func(vals ...interface{}) {
	name := "<UNKNOWN>"
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		name = runtime.FuncForPC(pc).Name()
		if i := strings.LastIndexByte(name, '.'); i != -1 {
			name = name[i + 1:]
		}
	}
	args := traceJoin(false, vals)
	return func(vals ...interface{}) {
		form := ""
		if len(vals) == 1 {
			form = "%-11v(%v) = %v"
		} else {
			form = "%-11v(%v) = (%v)"
		}
		rslt := ""
		rcvr := recover()
		if nil != rcvr {
			rslt = fmt.Sprintf("!PANIC:%v", rcvr)
		} else {
			rslt = traceJoin(true, vals)
		}
		log.Printf(form, name, args, rslt)
		if nil != rcvr {
			panic(rcvr)
		}
	}
}
