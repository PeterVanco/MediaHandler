package api

import (
	"MediaHandler/constants"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetSingle(c *gin.Context, root string) {
	var duplicates []string
	var duplicateUid string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if match := constants.DuplicateRegExp.FindString(info.Name()); match != "" {
			if strings.Contains(path, "thumb") {
				return filepath.SkipDir
			}
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
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(duplicates) == 0 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, duplicates)
}
