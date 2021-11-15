package geecache

type ByteView []byte

func (v ByteView) Len() int {
	return len(v)
}

func (v ByteView) ByteSlice() ByteView {
	c := make([]byte, len(v))
	copy(c, v)
	return c
}
