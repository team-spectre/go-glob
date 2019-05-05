package glob

import (
	"fmt"
)

type segmentType byte

const (
	literalSegment segmentType = iota
	runeMatchSegment
	questionSegment
	starSegment
	doubleStarSegment
	doubleStarSlashSegment
)

var segmentTypeNames = []string{
	"literalSegment",
	"runeMatchSegment",
	"questionSegment",
	"starSegment",
	"doubleStarSegment",
	"doubleStarSlashSegment",
}

func (x segmentType) String() string {
	if uint(x) >= uint(len(segmentTypeNames)) {
		return fmt.Sprintf("%%!segmentType(%d)", x)
	}
	return segmentTypeNames[x]
}

func (x segmentType) GoString() string {
	if uint(x) >= uint(len(segmentTypeNames)) {
		return fmt.Sprintf("segmentType(%d)", x)
	}
	return segmentTypeNames[x]
}

type parseState byte

const (
	rootState parseState = iota
	rootEscState
	rootOctState
	rootHexState
	charsetInitialState
	charsetHeadState
	charsetHeadEscState
	charsetHeadOctState
	charsetHeadHexState
	charsetMidState
	charsetTailState
	charsetTailEscState
	charsetTailOctState
	charsetTailHexState
)

var parseStateNames = []string{
	"rootState",
	"rootEscState",
	"rootOctState",
	"rootHexState",
	"charsetInitialState",
	"charsetHeadState",
	"charsetHeadEscState",
	"charsetHeadOctState",
	"charsetHeadHexState",
	"charsetMidState",
	"charsetTailState",
	"charsetTailEscState",
	"charsetTailOctState",
	"charsetTailHexState",
}

func (x parseState) String() string {
	if uint(x) >= uint(len(parseStateNames)) {
		return fmt.Sprintf("%%!parseState(%d)", x)
	}
	return parseStateNames[x]
}

func (x parseState) GoString() string {
	if uint(x) >= uint(len(parseStateNames)) {
		return fmt.Sprintf("parseState(%d)", x)
	}
	return parseStateNames[x]
}
