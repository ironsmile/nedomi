// +build go1.5

package pprof

import "net/http/pprof"

func init() {
	prefixToHandler["trace"] = pprof.Trace
}
