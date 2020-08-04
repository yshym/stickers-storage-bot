// Package db provides database operations and data structures
package db

// Sticker provides info about sticker
type Sticker struct {
	UserID       int
	FileUniqueID string
	FileID       string
}
