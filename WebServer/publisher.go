// publisher.go
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get array of songs
// Return string that represent it
func strSongsList(songsList []Song) string {

	// Check if the array is empty
	if len(songsList) == 0 {
		return "No results found."
	} else {
		// Prepare return result
		var stringOut string
		// For every song
		for _, song := range songsList {
			stringOut = stringOut + song.toString()
		}
		return stringOut
	}
}

// Define API functions
func setupRouter() *gin.Engine {

	r := gin.Default()

	// Ping test route
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Search song route
	r.GET("/search/songs/:songName", func(c *gin.Context) {

		// need to implement search in redis by text
		// meanwhile init the array
		songsFound := []Song{
			// songName, artistName, releaseYear, albumName, songLengthSec
			{001, "Shney Meshugaim", "Omer Adam", 2022, "Haim Shely", 239},
			{002, "Knafaim", "Tuna", 2019, "Abam", 211},
			{003, "Yesh Li Otach", "Moshe Peretz", 2016, "Love", 226},
		}

		for _, song := range songsFound {
			fmt.Printf("Title: %s, Artist: %s, releaseYear: %d\n", song.songName, song.artistName, song.releaseYear)
		}

		// Convert to text
		// msg := strSongsList(songsFound []Song)
		// println(strSongsList(songsFound))
		c.String(http.StatusOK, strSongsList(songsFound))
	})

	// Download song by id route
	r.GET("/download/songs/:id", func(c *gin.Context) {
		//idToDownload := c.Params.ByName("id")
		// Download song to client
		c.String(http.StatusOK, SimpleHello())
	})

	return r
}

func main() {
	// r := setupRouter()
	// Listen Server in 0.0.0.0:8080
	// r.Run(":8080")

	redisClient := ConnectRedisDB()

	songsList := []Song{
		// ID, songName, artistName, releaseYear, albumName, songLengthSec
		{001, "Shney Meshugaim", "Omer Adam", 2022, "Haim Shely", 239},
		{002, "Knafaim", "Tuna", 2019, "Abam", 211},
		{003, "Yesh Li Otach", "Moshe Peretz", 2016, "Love", 226},
	}

	err := AppendNewSongs(redisClient, songsList)
	if err != nil {
		fmt.Printf("Failed to append new songs to redis DB.")
	}

	redisClient.Close()
}
