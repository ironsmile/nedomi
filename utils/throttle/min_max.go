package throttle

import "time"

func min64(l, r int64) int64 {
	if l > r {
		return r
	}
	return l
}

func max64(l, r int64) int64 {
	if l > r {
		return l
	}
	return r
}

func maxDur(l, r time.Duration) time.Duration {
	if l > r {
		return l
	}
	return r
}
