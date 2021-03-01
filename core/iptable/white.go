package iptable

import (
	"ava/core"
)

func Allow(dst string) bool {
	if len(core.Sites) == 0 {
		return true
	}

	if !core.StringInSlice(dst, core.Sites) {
		return false
	}
	return true
}
