package guts

const (
	UintMax = ^uint(0)
	U64Max  = ^uint64(0)
	IntMax  = int(UintMax >> 1)
	IntMin  = ^IntMax
)
