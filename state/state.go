package state

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/google/renameio"
)

type FileState struct {
	filename string
}

func New(filename string) *FileState {
	return &FileState{filename: filename}
}

func (f *FileState) GetBytesToSkip() int64 {
	ff, err := loadState(f.filename)
	if err != nil {
		return 0
	}
	return ff.SkipBytes
}

func (f *FileState) GetInode() uint {
	ff, err := loadState(f.filename)
	if err != nil {
		return 0
	}
	return ff.Inode
}

func (f *FileState) Set(bytesToSkip int64, inode uint) error {
	return saveState(f.filename, State{SkipBytes: bytesToSkip, Inode: inode})
}

type State struct {
	SkipBytes int64 `json:"skip_bytes"`
	Inode     uint  `json:"inode"`
}

var (
	ErrStateFileCorrupted = errors.New("state file is corrupted")
)

func loadState(filename string) (*State, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var state State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, ErrStateFileCorrupted
	}
	return &state, nil
}

func saveState(filename string, state State) error {
	t, err := renameio.TempFile("", filename)
	if err != nil {
		return err
	}
	defer t.Cleanup()

	if err := json.NewEncoder(t).Encode(state); err != nil {
		return err
	}
	return t.CloseAtomicallyReplace()
}
