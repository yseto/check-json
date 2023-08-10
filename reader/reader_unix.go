//go:build !windows

package reader

import (
	"os"
	"syscall"
)

// https://github.com/mackerelio/go-check-plugins/blob/d59cfeeb8e33a8e8c7e76a107b3b862253abb5c2/check-log/lib/check-log_unix.go#L14
func detectInode(fi os.FileInfo) uint {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		return uint(stat.Ino)
	}
	return 0
}
