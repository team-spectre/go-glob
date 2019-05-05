package glob

const (
	uintMax = ^uint(0)
	u64Max  = ^uint64(0)
	intMax  = int(uintMax >> 1)
	intMin  = ^intMax
)
