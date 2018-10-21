package api

import (
	"MediaHandler/constants"
	. "MediaHandler/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetList(c *gin.Context, root string) {

	Logger.Printf("walking photo root %s\n", root)
	events := make(map[string]map[string]int)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		Logger.Printf("walkRoot found %s (%s)", info.Name(), path)
		if info.Name() == root {
			return nil
		}
		if strings.Contains(path, "Recycle") {
			return filepath.SkipDir
		}
		if info.IsDir() && path != root {
			Logger.Printf("walking event %s\n", path)
			events[path] = walkEvent(path)
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, events)
}
func walkEvent(path string) map[string]int {
	var duplicates []string
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// Logger.Printf("walkEvent found %s (%s)", info.Name(), path)
		if strings.Contains(path, "thumb") {
			return filepath.SkipDir
		}
		if match := constants.DuplicateRegExp.FindString(info.Name()); match != "" {
			duplicates = append(duplicates, match[:len(match)-1])
		}
		return nil
	})

	counter := make(map[string]int)
	for _, dup := range duplicates {
		counter[dup]++
	}
	return counter
}
