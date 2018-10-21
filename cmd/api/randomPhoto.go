package api

import (
	"MediaHandler/constants"
	. "MediaHandler/util"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const photoListCacheKey = "photoList"

var photoCache = cache.New(24*time.Hour, 24*time.Hour)

func GetRandom(c *gin.Context, root string) {
	photoList, err := getPhotoList(root)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(photoList) < 1 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	for {
		randomPhoto := photoList[rand.Intn(len(photoList))]
		if _, err := os.Stat(randomPhoto); os.IsNotExist(err) {
			photoCache.Delete(photoListCacheKey)
			continue
		}

		// TODO: Cache files shown in last X hours ?
		//if err := photoCache.Add(photoListCacheKey, photos, cache.DefaultExpiration); err != nil {
		//	continue
		//}
		GetResize(c, randomPhoto)
		return
	}
}

func getPhotoList(root string) ([]string, error) {
	if photoList, found := photoCache.Get(photoListCacheKey); found {
		return photoList.([]string), nil
	}

	Logger.Printf("building photo list cache\n")
	var photos []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "thumb") {
			return filepath.SkipDir
		}

		ext := filepath.Ext(info.Name())
		if len(ext) > 1 {
			ext = ext[1:]
		}
		if _, extensionSupported := constants.PhotoExtensions[ext]; !extensionSupported {
			return nil
		}
		photos = append(photos, path)
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, err
	}

	Logger.Printf("built photo list cache with %d entries\n", len(photos))
	if err := photoCache.Add(photoListCacheKey, photos, cache.DefaultExpiration); err != nil {
		return nil, err
	}

	return photos, nil
}
