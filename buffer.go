package glob

import (
	"sync"
)

const kMinRuneSliceCap = 64

var runeSlicePool = sync.Pool{New: newRuneSlice}

func newRuneSlice() interface{} {
	return make([]rune, 0, kMinRuneSliceCap)
}

func takeRuneSlice() []rune {
	return runeSlicePool.Get().([]rune)
}

func giveRuneSlice(slice []rune) {
	if slice != nil && cap(slice) >= kMinRuneSliceCap {
		runeSlicePool.Put(slice[:0])
	}
}

func copyRunes(in []rune) []rune {
	out := make([]rune, len(in))
	copy(out, in)
	return out
}
