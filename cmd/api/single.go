package api

import (
	"MediaHandler/constants"
	. "MediaHandler/util"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func GetSingle(c *gin.Context, root string) {
	var duplicates []string
	var duplicateUid string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if match := constants.DuplicateRegExp.FindString(info.Name()); match != "" {
			if duplicateUid == "" {
				duplicates = append(duplicates, path)
				duplicateUid = match
			} else if constants.DuplicateRegExp.FindString(info.Name()) == duplicateUid {
				duplicates = append(duplicates, path)
			} else {
				return io.EOF
			}
		}
		return nil
	})

	if err != nil && err != io.EOF {
		Logger.Panic(err)
	}

	c.JSON(http.StatusOK, duplicates)
}
