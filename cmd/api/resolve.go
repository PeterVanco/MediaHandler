package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func GetResolve(c *gin.Context) {

	// GIN cannot bind JSON arrays as root elements
	var photos []Photo
	raw, _ := c.GetRawData()
	_ = json.Unmarshal(raw, &photos)

	renamePhotos(photos)
	deletePhotos(photos)

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func renamePhotos(photos []Photo) {
	var namingSuffix string
	if keepCount(photos) > 1 {
		namingSuffix = "_"
	}
	for _, photo := range photos {
		if photo.Keep {
			renamePhoto(photo, namingSuffix)
		}
	}
}

func renamePhoto(photo Photo, namingSuffix string) {
	path, filename := filepath.Split(photo.Url)
	if match := regexp.MustCompile("DUP_([0-9]+)_").FindString(filename); match != "" {
		groups := regexp.MustCompile(`DUP_(?P<Uid>[0-9]+)_(?P<Filename>.*)\.(?P<Ext>.*)`).FindStringSubmatch(filename)
		newFilename := fmt.Sprintf("%s%s.%s", groups[1], namingSuffix, groups[2])
		os.Rename(photo.Url, filepath.Join(path, newFilename))
	}
	//r := regexp.MustCompile(`(?P<Year>\d{4})-(?P<Month>\d{2})-(?P<Day>\d{2})`)
	//r := regexp.MustCompile(`DUP_(?P<Uid>[0-9]+)_(?P<Month>\d{2})-(?P<Day>\d{2})`)
	//fmt.Printf("%#v\n", r.FindStringSubmatch(`2015-05-27`))
}

func keepCount(photos []Photo) int {
	keepCount := 0
	for _, photo := range photos {
		if photo.Keep {
			keepCount++
		}
	}
	return keepCount
}

func deletePhotos(photos []Photo) {
	for _, photo := range photos {
		if !photo.Keep {
			deletePhoto(photo)
		}
	}
}

func deletePhoto(photo Photo) {
	os.Remove(photo.Url)
}
