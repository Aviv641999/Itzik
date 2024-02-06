// redis.go
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

/*
Redis MetaData Conventions:
	For representing struct we'll use HASHES:
		song:<song_id>
			e.g: HSET song:123 title "Song Title"
		album:<album_id>
			e.g: HSET album:456 title "Album Title"
	For representing relationships we'll use SETS
		album:<album_id>:songs <song_id> <song_id> <song_id>
			e.g: SADD album:456:songs 123 456 789

*/

const (
	RedisAddress   = "192.168.72.129:6379"
	SongsFile      = "SongsToAppend.txt"
	RandomRange    = 99999
	SongPrefix     = "song:*"
	ResultsPerScan = 5
)

// getAllKeys return array of string of all key names in redis DB.
// Parameters:
//   - r: The Redis client.
//
// Returns:
//   - []string: array of string of all key names
//   - error: if error eccur, the error description.
func getAllKeys(r *redis.Client) ([]string, error) {
	log.Println("Start getAllKeys")
	arrResult, errKeys := r.Keys(context.Background(), "*").Result()
	if errKeys != nil {
		return nil, errKeys
	}
	return arrResult, nil
}

// generateNewIdForKey find new id for use by key.
// Parameters:
//   - r: The Redis client.
//   - key: The key of the hash to check. (song/album)
//
// Returns:
//   - string: The new id for use. ("" - if didn't succeed)
func generateNewIdForKey(r *redis.Client, key string) string {
	// Flag up, hash exists until find another
	hashExists := int64(1)
	counter := 0
	for hashExists == int64(1) && counter < 50 {
		// Generate a random number
		randomNumber := rand.Intn(RandomRange)
		// Format <song/album>:<5digitsID>
		key = fmt.Sprintf("%s:%s", key, fmt.Sprintf("%05d", randomNumber))
		// Check if hash exists
		hashExists, errExists := r.Exists(context.Background(), key).Result()
		if errExists != nil {
			return ""
		}
		if hashExists == int64(0) {
			log.Printf("Hash with key %s does not exists in Redis.\n", key)
			return fmt.Sprintf("%05d", randomNumber)
		} else {
			log.Printf("Hash with key %s does not exist in Redis.\n", key)
		}
		counter++
	}
	return ""
}

// buildSongsArrFromFile read from file and build songs array by it.
// Parameters:
//
// Returns:
//   - []Song: The array songs from the file.
//   - error: if error eccur, the error description.
func buildSongsArrFromFile(r *redis.Client) ([]Song, error) {
	var songsArr []Song
	// Open src songs file
	f, errOpenFile := os.OpenFile(SongsFile, os.O_RDONLY, 0666)
	if errOpenFile != nil {
		log.Fatalf("error opening file: %v", errOpenFile)
		return nil, errOpenFile
	}
	defer f.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(f)

	// Iterate over each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		// Split the line by commas
		parts := strings.Split(line, ",")
		// Trim spaces from each substring
		for i, str := range parts {
			parts[i] = strings.TrimSpace(str)
		}
		// Converts int to string
		releaseYearNum, _ := strconv.Atoi(parts[2])
		songLengthSecNum, _ := strconv.Atoi(parts[4])
		// Find new id to use for song
		id := generateNewIdForKey(r, "song")
		if id == "" {
			return nil, errors.New("Failed to generate new id for song.")
		}
		// Append new song by line data, format:
		// id, songName, artistName, releaseYear, albumName, songLengthSec
		songsArr = append(songsArr, Song{id: id, songName: parts[0], artistName: parts[1], releaseYear: releaseYearNum, albumName: parts[3], songLengthSec: songLengthSecNum})
		//log.Println("buildSongsArrFromFile:", id, parts[0], parts[1], releaseYearNum, parts[3], songLengthSecNum)
	}

	// Check for errors during scanning
	if errScanner := scanner.Err(); errScanner != nil {
		log.Fatalf("Error reading file: ", errScanner)
		return nil, errScanner
	}
	return songsArr, nil
}

// ConnectRedisDB make the initial connection with redisDB.
// Returns:
//   - r: The Redis client.
//   - error: if error eccur, the error description.
func ConnectRedisDB() (*redis.Client, error) {
	// Define redis client options
	r := redis.NewClient(&redis.Options{
		Addr:     RedisAddress, // Default Redis address
		Password: "",           // No password
		DB:       0,            // Default DB
	})

	// Ping the Redis server to check the connection
	pong, err := r.Ping(context.Background()).Result()
	if err != nil {
		log.Println("Error connecting to Redis:", err)
		return nil, err
	}
	log.Println("Connected to Redis:", pong)
	return r, nil
}

// setSongHash write to redis song HASH.
// Parameters:
//   - r: The Redis client.
//   - key: The redis key to write. (song:<5digitID>)
//   - songMap: map[string]string object that describe the song data.
//
// Returns:
//   - error: if error eccur, the error description.
func setSongHash(r *redis.Client, key string, songMap map[string]string) error {

	// Use HSet to set multiple fields in the hash
	err := r.HSet(context.Background(), key, songMap).Err()
	if err != nil {
		log.Printf("Failed to set multiple fields in hash: %s.", key)
		return err
	} else {
		log.Printf("Set successfully multiple fields in hash: %s.", key)
	}

	return nil
}

// AppendNewSongs write to RedisDB songs from file.
// Parameters:
//
// Returns:
//   - error: If succeeded return nil, If an error occurred returns error.
func AppendNewSongsRedis() error {
	// Start
	log.Println("Starting AppendNewSongs()")

	// Connect to RedisDB
	r, errConnectDB := ConnectRedisDB()
	if errConnectDB != nil {
		return errConnectDB
	}

	// Read from file and build songs array
	songsToAppend, errBuildArr := buildSongsArrFromFile(r)
	if errBuildArr != nil || songsToAppend == nil {
		return errors.New("Failed to build songs array from file.")
	}

	// Write songs metadata to RedisDB
	for _, song := range songsToAppend {
		songMap := song.toMap()
		key := "song:" + fmt.Sprint(song.id)
		setSongHash(r, key, songMap)
	}

	// End
	log.Println("AppendNewSongs: Close RedisDB")
	r.Close()
	return nil
}

// FlushAllRedis remove all data from RedisDB .
// Parameters:
//
// Returns:
//   - error: If succeeded return nil, If an error occurred returns error.
func FlushAllRedis() error {
	// Start
	log.Println("Starting FlushAllRedis()")

	// Connect to RedisDB
	r, errConnectDB := ConnectRedisDB()
	if errConnectDB != nil {
		return errConnectDB
	}

	// Use HSet to set multiple fields in the hash
	errFlush := r.FlushAll(context.Background()).Err()
	if errFlush != nil {
		log.Printf("Failed to remove all data in redis DB: %s.", errFlush)
		return errFlush
	}

	// End
	log.Println("FlushAllRedis: Close RedisDB")
	r.Close()
	return nil
}

// containsText check if the text contains in the array.
// Parameters:
//   - textToSearch: The text to search.
//   - slice: array of string to search in.
//
// Returns:
//   - bool: true if contains, false if not.
func containsText(songMap map[string]string, text string) bool {

	lowerText := strings.ToLower(text)
	for _, value := range songMap {
		if strings.Contains(strings.ToLower(value), lowerText) {
			return true
		}
	}
	return false

}

// mapToSong convert map to song struct.
// Parameters:
//   - songMap: represent song metadata from redis DB.
//
// Returns:
//   - song: true if contains, false if not.
func mapToSong(songMap map[string]string) (*Song, error) {
	song := &Song{}
	for key, value := range songMap {
		switch key {
		case "id":
			song.id = value
		case "songName":
			song.songName = value
		case "artistName":
			song.artistName = value
		case "albumName":
			song.albumName = value
		case "releaseYear":
			// Assuming the releaseYear is provided as string, needs conversion
			releaseYear, errYear := strconv.Atoi(value)
			if errYear != nil {
				log.Println("Error converting releaseYear:", errYear)
				return song, errYear
			}
			song.songLengthSec = releaseYear
		case "songLengthSec":
			// Assuming the songLengthSec is provided as string, needs conversion
			songLengthSec, errLength := strconv.Atoi(value)
			if errLength != nil {
				log.Println("Error converting songLengthSec:", errLength)
				return song, errLength
			}
			song.songLengthSec = songLengthSec
		}
	}
	if song.IsAllFilled() {
		return song, nil
	}
	return song, errors.New("Without full metadata of song.")

}

// SearchFreeText search in RedisDB songs/albums by free text .
// Parameters:
//   - textToSearch: The text to search.
//
// Returns:
//   - results: String that represents the result of search
//   - error: If succeeded return nil, If an error occurred returns error.
func SearchFreeText(textToSearch string) (string, error) {
	// Start
	log.Printf("Starting SearchFreeText(%s)", textToSearch)
	stringOut := ""

	// Connect to RedisDB
	r, errConnectDB := ConnectRedisDB()
	if errConnectDB != nil {
		return "", errConnectDB
	}

	// Get all keys
	// Set up the SCAN command with matching and counting
	cursor := uint64(0)

	// Until there is no results for Scan -> songs
	for {
		// Perform the SCAN command to get all keys
		keys, nextCursor, errScanKeys := r.Scan(context.Background(), cursor, SongPrefix, int64(ResultsPerScan)).Result()
		if errScanKeys != nil {
			log.Fatalf("Failed to SCAN keys: %v", errScanKeys)
			return "", errScanKeys
		}

		// Process the keys
		for _, key := range keys {
			// Retrieve the value of a specific fields in the hash
			value, err := r.HGetAll(context.Background(), key).Result()
			if err != nil {
				log.Printf("Error retrieving fields for key '%s': %v", key, err)
				continue
			}

			// Check if the value contains the desired text
			if containsText(value, textToSearch) {
				log.Printf("Key: %s, Value: %s\n", key, value)
				songToAdd, errMapToSong := mapToSong(value)
				if errMapToSong != nil {
					log.Printf("Error convert map to song '%s': %v", key, errMapToSong)
					continue
				}
				stringOut = stringOut + songToAdd.toString()
			}
		}

		// Update the cursor for the next iteration
		cursor = nextCursor

		// Check if the iteration is complete
		if cursor == 0 {
			break
		}
	}

	// End
	log.Println("SearchFreeText: Close RedisDB")
	r.Close()
	return stringOut, nil
}

func TestGenerate() {
	log.Println("Starting Test()")
	r, errConnectDB := ConnectRedisDB()
	if errConnectDB != nil {
		return
	}
	for i := 0; i < 5; i++ {
		fmt.Println("i: %d - ", i, generateNewIdForKey(r, "song"))
	}
	r.Close()
	log.Println("Ending TestGenerate")
}
