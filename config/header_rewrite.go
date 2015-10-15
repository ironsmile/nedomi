package config

import "github.com/ironsmile/nedomi/utils"

// HeadersRewrite rewrites headers
type HeadersRewrite struct {
	AddHeaders    HeaderPairs `json:"add_headers"`
	SetHeaders    HeaderPairs `json:"set_headers"`
	RemoveHeaders StringSlice `json:"remove_headers"`
}

// Copy returns a deep copy of the HeadersRewrite
func (h HeadersRewrite) Copy() HeadersRewrite {
	return HeadersRewrite{
		RemoveHeaders: utils.CopyStringSlice(h.RemoveHeaders),
		AddHeaders:    h.AddHeaders.Copy(),
		SetHeaders:    h.SetHeaders.Copy(),
	}
}
