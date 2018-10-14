package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/pbnjay/clustering"
	"github.com/rivo/duplo"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var (
	extensions = map[string]func(io.Reader) (image.Image, error){
		"jpg":  jpeg.Decode,
		"jpeg": jpeg.Decode,
		"png":  png.Decode,
		"gif":  gif.Decode,
	}
	sensitivity = flag.Int("sensitivity", 100, "the sensitivity threshold (the lower, the better the match (can be negative))")
	path        = flag.String("path", ".", "the path to search the images")
	dryRun      = flag.Bool("dryrun", false, "only print found matches")
)

var (
	buf bytes.Buffer
	// logger = log.New(&buf, "logger: ", log.Lshortfile)
	logger = log.New(bufio.NewWriter(os.Stdout), "", log.Lshortfile)
	store  *duplo.Store
	bar    *mpb.Bar
)

func init() {
	store = duplo.New()
	logger.SetOutput(os.Stdout)
	initBar(bar)
}
func initBar(bar *mpb.Bar) {
	p := mpb.New(
		// override default (80) width
		mpb.WithWidth(64),
		// override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// override default 120ms refresh rate
		mpb.WithRefreshRate(180*time.Millisecond*10000),
	)

	name := "Processed Images:"
	bar = p.AddBar(0,
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			decor.OnComplete(
				// ETA decorator with ewma age of 60, and width reservation of 4
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 4}), "done",
			),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)
}

func main() {

	distanceMap := make(clustering.DistanceMap)

	flag.Parse()
	*sensitivity -= 100

	files, err := ioutil.ReadDir(*path)
	if err != nil {
		log.Fatal(err)
	}

	logger.Printf("found %d files\n", len(files))
	//bar.SetTotal(int64(len(files)), true)

	for _, f := range files {
		handleFile(f, distanceMap)
		// bar.Increment()
	}
	// p.Wait()

	logger.Printf("calculating clusters from %d hashes\n", len(store.IDs()))
	calculateClusters(clustering.NewDistanceMapClusterSet(distanceMap))
}

func calculateClusters(clusterSet clustering.ClusterSet) {
	clustering.Cluster(clusterSet, clustering.Threshold(float64(*sensitivity)), clustering.CompleteLinkage())
	// Enumerate clusters and print members
	clusterSet.EachCluster(-1, func(cluster int) {
		clusterSize := 0
		clusterSet.EachItem(cluster, func(x clustering.ClusterItem) {
			clusterSize = clusterSize + 1
		})

		if clusterSize > 1 {
			uidLen := 10
			uid := int64(math.Pow10(uidLen-1)) + rand.Int63n(int64(math.Pow10(uidLen)))
			logger.Printf("found duplicate set as cluster %d with size=%d, assigning uid=%d\n", cluster, clusterSize, uid)
			clusterSet.EachItem(cluster, func(item clustering.ClusterItem) {
				moveFile(item.(string), uid)
			})
		}
	})
}

func handleFile(f os.FileInfo, distanceMap clustering.DistanceMap) {
	ext := filepath.Ext(f.Name())
	if len(ext) > 1 {
		ext = ext[1:]
	}
	if _, ok := extensions[ext]; !ok {
		return
	}
	filename := filepath.Join(*path, f.Name())
	file, err := os.Open(filename)
	if err != nil {
		logger.Printf("%s: %v\n", filename, err)
		return
	}
	defer file.Close()

	// TODO: use image.Decode() ?
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		logger.Printf("%s: %v\n", filename, err)
		return
	}
	if decodeFunc, ok := extensions[format]; ok {
		if _, err = file.Seek(0, 0); err != nil {
			logger.Printf("could not seek (rewind) %s: %v\n", filename, err)
			return
		}
		img, err := decodeFunc(file)
		if err != nil {
			logger.Printf("ignoring %s: %v\n", filename, err)
			return
		}
		handleImage(filename, img, f, distanceMap)
	}
}

func handleImage(filename string, img image.Image, f os.FileInfo, distanceMap clustering.DistanceMap) {
	// Add image "img" to the store.
	logger.Printf("hashing %s\n", f.Name())
	hash, _ := duplo.CreateHash(img)
	matches := store.Query(hash)
	store.Add(f, hash)
	distanceItem := make(map[clustering.ClusterItem]float64)
	if len(matches) > 0 {
		sort.Sort(matches)

		for _, match := range matches {
			fi := match.ID.(os.FileInfo)
			logger.Printf("    %s score with %s (%d)\n", filename, fi.Name(), int(match.Score))
			distanceItem[fi.Name()] = match.Score
		}

		//match := matches[0]
		//fi := match.ID.(os.FileInfo)
		//// logger.Printf("%s closest score with %s (%d)\n", filename, fi.Name(), int(match.Score))
		//if int(match.Score) <= *sensitivity {
		//	logger.Printf("%s matches: %s\n", filename, fi.Name())
		//}
	} else {
		logger.Println("file has no matches ", filename)
	}
	distanceMap[f.Name()] = distanceItem
}

func moveFile(filename string, uid int64) {
	newFilename := fmt.Sprintf("%s_%d_%s", "DUP", uid, filename)
	if *dryRun {
		logger.Printf("    would move file %s to %s\n", filename, newFilename)
	} else {
		if err := os.Rename(filepath.Join(*path, filename), filepath.Join(*path, newFilename)); err != nil {
			// logger.Printf("error moving file %s to %s\n", filename, newFilename)
		}
	}
}
