package api

import (
	"MediaHandler/constants"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetSingle(c *gin.Context, root string) {

	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(info.Name(), constants.DuplicateMarker) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Println(file)
	}

	c.JSON(http.StatusOK, files)
}
