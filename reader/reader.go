package reader

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type RotateReader struct {
	filename string
	state    StateHolder
}

type StateHolder interface {
	GetBytesToSkip() int64
	GetInode() uint
	Set(bytesToSkip int64, inode uint) error
}

var errFileNotFoundByInode = errors.New("old file not found")

func New(filename string, state StateHolder) *RotateReader {
	return &RotateReader{filename: filename, state: state}
}

func (rr *RotateReader) Read(ctx context.Context, callback func(io.Reader)) error {
	var skipBytes = rr.state.GetBytesToSkip()
	var inode = rr.state.GetInode()

	f, err := os.Open(rr.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// read old rotated file
	var oldf *os.File
	oldf, err = openOldFile(rr.filename, skipBytes, inode)
	if err != nil {
		return err
	}
	defer oldf.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if stat.Size() < skipBytes {
		// file is rotated.
	} else if skipBytes > 0 {
		f.Seek(skipBytes, io.SeekStart)
	}

	if oldf != nil {
		callback(io.MultiReader(oldf, f))
	} else {
		callback(f)
	}

	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	return rr.state.Set(offset, detectInode(stat))
}

// refer https://github.com/mackerelio/go-check-plugins/blob/d59cfeeb8e33a8e8c7e76a107b3b862253abb5c2/check-log/lib/check-log.go#L574
func openOldFile(f string, skipBytes int64, oldInode uint) (*os.File, error) {
	fi, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	inode := detectInode(fi)
	if inode > 0 && oldInode != inode {
		if oldFile, err := findFileByInode(oldInode, filepath.Dir(f)); err == nil {
			oldf, err := os.Open(oldFile)
			if err != nil {
				return nil, err
			}
			oldfi, _ := oldf.Stat()
			if oldfi.Size() > skipBytes {
				oldf.Seek(skipBytes, io.SeekStart)
				return oldf, nil
			}
		} else if err != errFileNotFoundByInode {
			return nil, err
		}
		// just ignore the process of searching old file if errFileNotFoundByInode
	}
	return nil, nil
}

// refer https://github.com/mackerelio/go-check-plugins/blob/master/check-log/lib/check-log.go#L557
func findFileByInode(inode uint, dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			return "", err
		}
		if detectInode(fi) == inode {
			return filepath.Join(dir, fi.Name()), nil
		}
	}
	return "", errFileNotFoundByInode
}
