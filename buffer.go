package glob

import (
	"sync"
)

const (
	kMinByteSliceCap = 256
	kMinRuneSliceCap = 64
)

var (
	byteSlicePool = sync.Pool{New: newByteSlice}
	runeSlicePool = sync.Pool{New: newRuneSlice}
)

func newByteSlice() interface{} {
	return make([]byte, 0, kMinByteSliceCap)
}

func takeByteSlice(n uint) []byte {
	if n > kMinByteSliceCap {
		return make([]byte, 0, n)
	}
	return byteSlicePool.Get().([]byte)
}

func giveByteSlice(slice []byte) {
	if slice != nil && cap(slice) >= kMinByteSliceCap {
		byteSlicePool.Put(slice[:0])
	}
}

func newRuneSlice() interface{} {
	return make([]rune, 0, kMinRuneSliceCap)
}

func takeRuneSlice(n uint) []rune {
	if n > kMinRuneSliceCap {
		return make([]rune, 0, n)
	}
	return runeSlicePool.Get().([]rune)
}

func giveRuneSlice(slice []rune) {
	if slice != nil && cap(slice) >= kMinRuneSliceCap {
		runeSlicePool.Put(slice[:0])
	}
}

func copyRunes(in []rune) []rune {
	n := uint(len(in))
	if n == 0 {
		return nil
	}
	out := make([]rune, n)
	copy(out, in)
	return out
}
