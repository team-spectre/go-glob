//
// Package glob provides advanced filesystem-independent glob pattern matching.
//
// Typical usage might be to match a single glob pattern against multiple paths:
//
//	g := MustCompile(`/some/dir/**/deep/subdir/*.ext`)
//	for _, path := range paths {
//		if g.Matcher(path).Matches() {
//			fmt.Println(path)
//		}
//	}
//
// The Matcher type also provides an iterative interface for progressively
// matching a path, allowing access to more detailed match information via the
// Capture type.
//
package glob // import "github.com/team-spectre/go-glob"
