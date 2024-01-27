// redis.go
package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	RedisAddress = "192.168.72.129:6379"
)

func printMap(m map[string]string) {
	// loop over keys and values in the map.
	for k, v := range m {
		fmt.Println("print map", k, "value is", v)
	}
}

func SimpleHello() string {
	return fmt.Sprintf("Hey from redis.go")
}

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
		fmt.Println("Error connecting to Redis:", err)
		return nil
	}
	fmt.Println("Connected to Redis:", pong)
	return r
}

// RedisExample is a function that demonstrates Redis-related operations
func RedisExample() {

	var r = ConnectRedisDB()
	// Close the Redis connection when done
	defer r.Close()
	// If connected successfully
	if r != nil {

		// Set a key-value pair in Redis
		err := r.Set(context.Background(), "hello", "world", 0).Err()
		if err != nil {
			fmt.Println("Error setting key in Redis:", err)
			return
		}

		// Get the value of the key from Redis
		val, err := r.Get(context.Background(), "hello").Result()
		if err != nil {
			fmt.Println("Error getting key from Redis:", err)
			return
		}
		fmt.Printf("Key: hello, Value: %s\n", val)
	} else {
		// When error occurred while connecting to Redis
		return
	}

}

// setSongHash sets the Redis hash for a song
func setSongHash(r *redis.Client, key string, songMap map[string]string) error {

	// Use HSet to set multiple fields in the hash
	err := r.HSet(context.Background(), key, songMap).Err()
	if err != nil {
		fmt.Printf("Failed to set multiple fields in hash: %s.", key)
		return err
	} else {
		fmt.Printf("Set successfully multiple fields in hash: %s.", key)
	}

	return nil
}

// Fill redis db in songs hashes
func AppendNewSongs(r *redis.Client, songsList []Song) error {

	for _, song := range songsList {
		songMap := song.toMap()
		setSongHash(r, "song:"+fmt.Sprint(song.id), songMap)
	}
	return nil
}
