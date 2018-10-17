package api

import (
	"MediaHandler/constants"
	. "MediaHandler/util"
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
	if match := constants.DuplicateRegExp.FindString(filename); match != "" {
		groups := regexp.MustCompile(`DUP_(?P<Uid>[0-9]+)_(?P<Filename>.*)\.(?P<Ext>.*)`).FindStringSubmatch(filename)
		newFilename := fmt.Sprintf("%s%s.%s", groups[2], namingSuffix, groups[3])
		if dryRun, _ := Flags.GetBool("dryrun"); dryRun {
			Logger.Printf("    would rename file %s to %s\n", photo.Url, filepath.Join(path, newFilename))
		} else {
			Logger.Printf("    renaming file %s to %s\n", photo.Url, filepath.Join(path, newFilename))
			if err := os.Rename(photo.Url, filepath.Join(path, newFilename)); err != nil {
				Logger.Printf("    ! could not rename file %s to %s\n", photo.Url, filepath.Join(path, newFilename))
			}
		}
	}
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
	if dryRun, _ := Flags.GetBool("dryrun"); dryRun {
		Logger.Printf("    would delete file %s\n", photo.Url)
	} else {
		Logger.Printf("    deleting file %s\n", photo.Url)
		if err := os.Remove(photo.Url); err != nil {
			Logger.Printf("    ! could not delete file %s\n", photo.Url)
		}
	}
}
