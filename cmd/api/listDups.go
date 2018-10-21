package api

import (
	"MediaHandler/constants"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type SameSizeMap map[int64][]string

func GetDuplicates(c *gin.Context, root string) {
	photoMap, err := getPhotoMap(root)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(photoMap) < 1 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	filterSingleFiles(photoMap)

	c.JSON(http.StatusOK, photoMap)
}

func filterSingleFiles(photoMap map[string]SameSizeMap) {
	for fileName, sameSizeMap := range photoMap {
		for size, pathList := range sameSizeMap {
			if len(pathList) < 2 {
				delete(sameSizeMap, size)
			}
		}
		if len(sameSizeMap) == 0 {
			delete(photoMap, fileName)
		}
	}
}

func getPhotoMap(root string) (map[string]SameSizeMap, error) {
	photoMap := make(map[string]SameSizeMap)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if constants.FilterPAth(path) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}
		if photoMap[info.Name()] == nil {
			photoMap[info.Name()] = make(SameSizeMap)
		}
		photoMap[info.Name()][info.Size()] = append(photoMap[info.Name()][info.Size()], path)
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, err
	}

	return photoMap, nil
}
