package models

import (
	"fmt"
	"time"
)

type Track struct {
	Id         int64
	Title      string
	Album      string
	Artist     string
	Composer   string
	Genre      string
	Year       int
	Length     time.Duration
	Bitrate    int
	Channels   int
	Samplerate int
	Path       string
}

func (t Track) String() string {
	return fmt.Sprintf("{ Title: %v, Album: %v, Composer: %v, Genre: %v, Year: %v, Length: %v, Bitrate: %v, Channels: %v, Samplerate: %v, Path: %v}",
		t.Title, t.Album, t.Composer, t.Genre, t.Year, t.Length, t.Bitrate, t.Channels, t.Samplerate, t.Path)
}
