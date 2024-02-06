// publisher.go
package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	LogFile = "log.txt"
)

// strSongsList return string that represents array of songs.
// Parameters:
//   - songsList: []Song array of song data.
//
// Returns:
//   - string: represents array of songs
//   - error: If succeeded return nil, If an error occurred returns error.
func strSongsList(songsList []Song) (string, error) {

	// Check if the array is empty
	if len(songsList) == 0 {
		return "", errors.New("The array is empty.")
	}
	// Prepare return result
	var stringOut string
	// For every song
	for _, song := range songsList {
		stringOut = stringOut + song.toString()
	}
	return stringOut, nil
}

// Define API functions
func setupRouter() *gin.Engine {

	r := gin.Default()

	// Ping test route
	r.GET("/ping", func(c *gin.Context) {
		// Start
		log.Println("Ping!")
		// End
		c.String(http.StatusOK, "pong")
	})

	// Search route
	r.GET("/search/:textToSearch", func(c *gin.Context) {
		// Start
		log.Println("Search!")
		textToSearch := c.Params.ByName("textToSearch")

		// Search text in RedisDB
		stringOut, errSongsArray := SearchFreeText(textToSearch)
		if errSongsArray != nil || stringOut == "" {
			c.String(http.StatusOK, "There is no results.")
		}

		//End
		c.String(http.StatusOK, stringOut)
	})

	// Download song by id route
	r.GET("/download/songs/:id", func(c *gin.Context) {
		log.Println("Download!")
		//idToDownload := c.Params.ByName("id")
		// Download song to client
		c.String(http.StatusOK, "Download?")
	})

	// Initilization of RedisDB
	r.GET("/redis/addsongs", func(c *gin.Context) {
		// Start
		log.Println("Add songs!")

		// Write songs metadata to RedisDB
		errAppendSongs := AppendNewSongsRedis()
		if errAppendSongs != nil {
			log.Printf("Failed to append new songs to redis DB: %v", errAppendSongs)
			c.String(http.StatusOK, "Failed to append new songs to redis DB.")
		}

		// End
		c.String(http.StatusOK, "pong")
	})

	// Flush all data in RedisDB
	r.GET("/redis/flushall", func(c *gin.Context) {
		// Start
		log.Println("Flush all!")

		// Remove all data in Redis DB
		errFlushAll := FlushAllRedis()
		if errFlushAll != nil {
			log.Printf("Failed to delete all data in Redis DB: %v", errFlushAll)
			c.String(http.StatusOK, "Failed to delete all data in Redis DB.")
		}

		// End
		c.String(http.StatusOK, "pong")
	})

	return r
}

func main() {
	// Setting log mechanism
	f, errOpenFile := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errOpenFile != nil {
		log.Fatalf("error opening file: %v", errOpenFile)
	}
	log.SetOutput(f)
	defer f.Close()
	defer log.Println("Ending main")

	log.Println("--- Starting main... ---")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//End
	r := setupRouter()
	log.Println("Server is listening in 0.0.0.0:8080")
	r.Run(":8080")

}
