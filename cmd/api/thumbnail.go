package api

import (
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"image"
	"image/jpeg"
	"net/http"
	"os"
)

func GetResize(c *gin.Context, path string) {
	file, err := os.Open(path)
	if err != nil {
		c.Error(err)
		c.String(http.StatusNotFound, "")
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		c.Error(err)
		return
	}
	file.Close()

	resizedImg := imaging.Resize(img, 500, 0, imaging.Lanczos)
	jpeg.Encode(c.Writer, resizedImg, nil)
}
