package models

import "time"

type Track struct {
	Title       string
	Album       string
	AlbumArtist string
	Composer    string
	Genre       string
	Year        int
	Length      time.Duration
}
