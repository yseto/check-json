package state

type EmptyState struct{}

func Empty() *EmptyState {
	return &EmptyState{}
}

func (f *EmptyState) GetBytesToSkip() int64 {
	return 0
}

func (f *EmptyState) GetInode() uint {
	return 0
}

func (f *EmptyState) Set(bytesToSkip int64, inode uint) error {
	return nil
}
