package guts

import (
	"fmt"
)

const (
	LiteralSegment SegmentType = iota
	RuneMatchSegment
	QuestionSegment
	StarSegment
	DoubleStarSegment
	DoubleStarSlashSegment
)

var segmentTypeNames = []string{
	"LiteralSegment",
	"RuneMatchSegment",
	"QuestionSegment",
	"StarSegment",
	"DoubleStarSegment",
	"DoubleStarSlashSegment",
}

func (x SegmentType) String() string {
	if uint(x) >= uint(len(segmentTypeNames)) {
		return fmt.Sprintf("%%!SegmentType(%d)", x)
	}
	return segmentTypeNames[x]
}

func (x SegmentType) GoString() string {
	if uint(x) >= uint(len(segmentTypeNames)) {
		return fmt.Sprintf("SegmentType(%d)", x)
	}
	return segmentTypeNames[x]
}

const (
	RootState ParseState = iota
	RootEscState
	RootOctState
	RootHexState
	CharsetInitialState
	CharsetHeadState
	CharsetHeadEscState
	CharsetHeadOctState
	CharsetHeadHexState
	CharsetMidState
	CharsetTailState
	CharsetTailEscState
	CharsetTailOctState
	CharsetTailHexState
)

var parseStateNames = []string{
	"RootState",
	"RootEscState",
	"RootOctState",
	"RootHexState",
	"CharsetInitialState",
	"CharsetHeadState",
	"CharsetHeadEscState",
	"CharsetHeadOctState",
	"CharsetHeadHexState",
	"CharsetMidState",
	"CharsetTailState",
	"CharsetTailEscState",
	"CharsetTailOctState",
	"CharsetTailHexState",
}

func (x ParseState) String() string {
	if uint(x) >= uint(len(parseStateNames)) {
		return fmt.Sprintf("%%!ParseState(%d)", x)
	}
	return parseStateNames[x]
}

func (x ParseState) GoString() string {
	if uint(x) >= uint(len(parseStateNames)) {
		return fmt.Sprintf("ParseState(%d)", x)
	}
	return parseStateNames[x]
}
