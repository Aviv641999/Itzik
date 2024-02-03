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
	RedisAddress = "192.168.72.129:6379"
	SongsFile    = "SongsToAppend.txt"
	RandomRange  = 99999
)

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
//   - r: The Redis client.
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
func ConnectRedisDB() *redis.Client {
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
		return nil
	}
	log.Println("Connected to Redis:", pong)
	return r
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

// AppendNewSongs write to redis songs from .
// Parameters:
//   - r: The Redis client.
//   - key: The redis key to write. (song:<5digitID>)
//   - songMap: map[string]string object that describe the song data.
//
// Returns:
//   - error: if error eccur, the error description.
func AppendNewSongs() error {

	log.Println("Starting AppendNewSongs()")
	r := ConnectRedisDB()

	songsToAppend, errBuildArr := buildSongsArrFromFile(r)
	if errBuildArr != nil {
		return errBuildArr
	}
	if songsToAppend == nil {
		return errors.New("Failed to build songs array from file.")
	}
	for _, song := range songsToAppend {
		songMap := song.toMap()
		key := "song:" + fmt.Sprint(song.id)
		setSongHash(r, key, songMap)
	}

	log.Println("Close RedisDB")
	r.Close()
	return nil
}
