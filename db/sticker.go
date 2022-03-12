// Package db provides database operations and data structures
package db

// Sticker provides info about sticker
type Sticker struct {
	UserID       int
	FileUniqueID string
	FileID       string
	Timestamp    string
	UseCount     uint
}

// Stickers provides a slice of stickers
type Stickers []Sticker

// Len returns a number of stickers
func (s Stickers) Len() int {
	return len(s)
}

// Less checks if a first sticker should be shown before a second one
func (s Stickers) Less(i, j int) bool {
	useCount1, useCount2 := s[i].UseCount, s[j].UseCount
	timestamp1, timestamp2 := s[i].Timestamp, s[j].Timestamp
	if useCount1 == 0 && useCount2 == 0 {
		return timestamp1 < timestamp2
	}
	return useCount1 > useCount2
}

// Swap swaps two stickers
func (s Stickers) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
