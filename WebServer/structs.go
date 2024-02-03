// redis.go
package main

import (
	"fmt"
	"runtime/debug"
	"strconv"
)

type Song struct {
	id            string
	songName      string
	artistName    string
	releaseYear   int
	albumName     string
	songLengthSec int
}

type Album struct {
	artistName  string
	releaseYear int
	songsList   []Song
}

// Method returns string that represent it
func (s Song) toString() string {
	return fmt.Sprintf("ID: %s, Name: %s, Artist: %s, Release Year: %d, Album: %s, Length: %d\n", s.id, s.songName, s.artistName, s.releaseYear, s.albumName, s.songLengthSec)
}

// Method convert song struct to map
func (s Song) toMap() map[string]string {

	// Error handeling
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Printf("%v, %s", panicInfo, string(debug.Stack()))
		}
	}()

	songMap := make(map[string]string)

	songMap["id"] = s.id
	songMap["name"] = s.songName
	songMap["artistName"] = s.artistName
	songMap["releaseYear"] = strconv.Itoa(s.releaseYear)
	songMap["albumName"] = s.albumName
	songMap["songLengthSec"] = strconv.Itoa(s.songLengthSec)

	return songMap
}
