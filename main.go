package main

import (
	"bufio"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"time"
)

//go:embed assets/* templates/*
var f embed.FS

func main() {

	router := gin.New()

	// Middleware!
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.TrustedPlatform = gin.PlatformCloudflare

	templ := template.Must(template.New("").ParseFS(f, "templates/*.tmpl"))
	router.SetHTMLTemplate(templ)

	// example: /public/assets/images/example.png
	router.StaticFS("/public", http.FS(f))

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})

	router.GET("favicon.ico", func(c *gin.Context) {
		file, _ := f.ReadFile("assets/favicon.ico")
		c.Data(
			http.StatusOK,
			"image/x-icon",
			file,
		)
	})

	router.POST("/submit", handleUpload)

	router.Run(":8080")
}

func handleUpload(c *gin.Context) {
	// This will come from a POST with hopefully a file that  we can process
	file, err := c.FormFile("gamelog")
	if err != nil {
		c.Error(err)
		return
	}

	log.Println(file.Filename)
	log.Println(file.Size)

	f, err := file.Open()
	if err != nil {
		c.Error(fmt.Errorf("failed to open file from header: %w", err))
		return
	}
	defer f.Close()

	ll := LogList{
		Name:      file.Filename,
		Character: "nil",
		Lines:     []GameLog{},
	}

	br := bufio.NewScanner(f)
	for br.Scan() {
		line := br.Text()
		if len(line) < 25 {
			continue
		}
		parseLogLine(&ll, line)
	}

	log.Printf("%#v", ll)

	types := make(map[string]map[string]int)

	startTime := time.Unix(1987654321, 0)
	endTime := time.Unix(0, 0)

	for _, log := range ll.Lines {
		if log.Timestamp.Before(startTime) {
			startTime = log.Timestamp
		}
		if log.Timestamp.After(endTime) {
			endTime = log.Timestamp
		}
		for _, j := range log.Minings {
			if types[j.Ore] == nil {
				types[j.Ore] = make(map[string]int)
			}
			types[j.Ore]["mined"] += j.Amount
			types[j.Ore]["wasted"] += j.Waste
		}
	}

	for ore, mp := range types {
		types[ore]["efficiency"] = (mp["mined"] * 100.0) / (mp["mined"] + mp["wasted"])
	}

	c.HTML(http.StatusOK, "parsed.tmpl", gin.H{
		"minings":  types,
		"logStart": startTime.Format(time.RFC1123),
		"logEnd":   endTime.Format(time.RFC1123),
		"elapsed":  endTime.Sub(startTime).String(),
	})

}
