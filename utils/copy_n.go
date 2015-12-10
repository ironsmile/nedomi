package utils

import "io"

// CopyN uses io.CopyN but tries to only have ONE io.LimitReader
func CopyN(w io.Writer, r io.Reader, limit int64) (n int64, err error) {
	if lr, ok := r.(*io.LimitedReader); ok {
		if lr.N <= limit {
			if llr, ok := lr.R.(*io.LimitedReader); ok && llr.N <= lr.N {
				return CopyN(w, lr.R, lr.N)
			}
			return io.Copy(w, r)
		}
		return CopyN(w, lr.R, limit)
	}
	return io.CopyN(w, r, limit)
}

func min64(l, r int64) int64 {
	if l > r {
		return r
	}
	return l
}
