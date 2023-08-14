package main

import (
	"bufio"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {

	r := gin.Default()

	// Ping test route
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Search song route
	r.GET("/search/all/:freetext", func(c *gin.Context) {
		textToSearch := c.Params.ByName("freetext")
		//textToSearch := c.Params.ByName("freetext")
		f, err := os.Open("C:\\Users\\Aviv\\Desktop\\Aviv\\Cyber\\איציק\\Code\\ExcelDB\\SongsMetadata.txt")
		if err != nil {
			c.String(http.StatusOK, "Do Nothing")
		}
		defer f.Close()

		// Splits on newlines by default.
		scanner := bufio.NewScanner(f)

		line := 1
		// https://golang.org/pkg/bufio/#Scanner.Scan
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), textToSearch) {
				c.String(http.StatusOK, "%s\n", strconv.Itoa(line))
			}

			line++
		}

		if err := scanner.Err(); err != nil {
			// Handle the error
		}

		msg := "Bob Dylan"

		c.String(http.StatusOK, msg)
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
