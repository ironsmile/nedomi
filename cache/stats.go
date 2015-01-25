package cache

import (
	"fmt"
	"github.com/gophergala/nedomi/config"
)

type CacheStats struct {
	ID       string
	Hits     uint64
	Requests uint64
	Size     config.BytesSize
	Objects  uint64
}

func (c *CacheStats) CachHitPrc() string {
	if c.Requests == 0 {
		return ""
	}
	return fmt.Sprintf("%.f%%", (float32(c.Hits)/float32(c.Requests))*100)
}
