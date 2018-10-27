package cmd

import (
	"MediaHandler/constants"
	. "MediaHandler/util"
	"fmt"
	"github.com/pbnjay/clustering"
	"github.com/rivo/duplo"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"image"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var findDuplicatesCmd *cobra.Command

var (
	// params
	sensitivity int

	// globals
	store *duplo.Store
	bar   *mpb.Bar
)

func init() {

	findDuplicatesCmd = &cobra.Command{
		Use:   "find-duplicates",
		Short: "finds duplicates in in defined path",
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			dryRun, _ := cmd.Flags().GetBool("dryrun")
			findDuplicates(path, dryRun)
		},
	}

	RootCmd.AddCommand(findDuplicatesCmd)
	findDuplicatesCmd.Flags().IntVar(&sensitivity, "sensitivity", 0, "the sensitivity threshold (the lower, the better the match (can be negative))")
	findDuplicatesCmd.Flags().String("path", ".", "the path to search the images")
	findDuplicatesCmd.Flags().Int("uidlen", 8, "length of duplicate UID")
}

func findDuplicates(path string, dryRun bool) {
	bar = initBar()
	store = duplo.New()

	files := findImages(path)
	Logger.Printf("found %d files\n", len(files))
	bar.SetTotal(int64(len(files)), false)

	distanceMap := addImagesToStore(files)

	for _, id := range store.IDs() {
		Logger.Printf("item %s is in the store\n", id.(string))
	}

	Logger.Printf("calculating clusters from %d hashes\n", len(store.IDs()))
	clusterSet := clustering.NewDistanceMapClusterSet(distanceMap)
	clustering.Cluster(clusterSet, clustering.Threshold(float64(sensitivity)), clustering.CompleteLinkage())
	moveClusteredDuplicates(clusterSet, dryRun)
}

func findImages(path string) []string {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		Logger.Println(path)
		if !constants.FilterPath(path) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil && err != io.EOF {
		panic(err)
	}
	return files
}

func addImagesToStore(files []string) clustering.DistanceMap {
	distanceMap := make(clustering.DistanceMap)
	for _, filePath := range files {
		handleFile(filePath, distanceMap)
		bar.Increment()
	}
	return distanceMap
}

func initBar() *mpb.Bar {
	p := mpb.New(
		// override default (80) width
		mpb.WithWidth(64),
		// override default "[=>-]" format
		mpb.WithFormat("╢▌▌░╟"),
		// override default 120ms refresh rate
		mpb.WithRefreshRate(10*time.Second),
	)

	//name := "Processed Images:"
	return p.AddBar(0,
		mpb.PrependDecorators(
			// display our name with one space on the right
			// decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
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

func moveClusteredDuplicates(clusterSet clustering.ClusterSet, dryRun bool) {
	clusterSet.EachCluster(-1, func(cluster int) {
		clusterSize := 0
		clusterSet.EachItem(cluster, func(x clustering.ClusterItem) {
			clusterSize = clusterSize + 1
		})

		if clusterSize > 1 {
			uidLen, _ := findDuplicatesCmd.Flags().GetInt("uidlen")
			uid := int64(math.Pow10(uidLen-1)) + rand.Int63n(int64(math.Pow10(uidLen)))
			Logger.Printf("found duplicate set as cluster %d with size=%d, assigning uid=%d\n", cluster, clusterSize, uid)
			clusterSet.EachItem(cluster, func(item clustering.ClusterItem) {
				moveFile(item.(string), uid, dryRun)
			})
		}
	})
}

func handleFile(filePath string, distanceMap clustering.DistanceMap) {
	ext := filepath.Ext(filePath)
	if len(ext) > 1 {
		ext = ext[1:]
	}

	if _, extensionSupported := constants.PhotoExtensions[ext]; !extensionSupported {
		return
	}

	// filename := filepath.Join(basePath, fileInfo.Name())
	file, err := os.Open(filePath)
	if err != nil {
		Logger.Printf("%s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		Logger.Printf("%s: %v\n", filePath, err)
		return
	}
	handleImage(filePath, img, distanceMap)
}

func handleImage(filePath string, img image.Image, distanceMap clustering.DistanceMap) {
	Logger.Printf("hashing %s\n", filePath)
	hash, _ := duplo.CreateHash(img)
	matches := store.Query(hash)
	store.Add(filePath, hash)
	distanceItem := make(map[clustering.ClusterItem]float64)
	if len(matches) > 0 {
		sort.Sort(matches)
		Logger.Printf("adding %d matches to cluster set\n", len(matches))

		for _, match := range matches {
			matchFilePath := match.ID.(string)
			Logger.Printf("    %s score with %s (%d)\n", filePath, matchFilePath, int(match.Score))
			distanceItem[matchFilePath] = match.Score
		}
	} else {
		Logger.Println("file has no matches ", filePath)
	}
	distanceMap[filePath] = distanceItem
}

func moveFile(filePath string, uid int64, dryRun bool) {

	fileDir, fileName := filepath.Split(filePath)
	filenameSuffix := fileName
	if match := constants.DuplicateRegExp.FindString(fileName); match != "" {
		filenameSuffix = strings.TrimPrefix(fileName, match)
	}
	newFilename := fmt.Sprintf("%s_%d_%s", "DUP", uid, filenameSuffix)
	if dryRun {
		Logger.Printf("    would move file %s to %s\n", fileName, newFilename)
	} else {
		Logger.Printf("moving file %s to %s\n", fileName, newFilename)
		if err := os.Rename(filePath, filepath.Join(fileDir, newFilename)); err != nil {
			Logger.Printf("error moving file %s to %s\n", fileName, newFilename)
		}
	}
}
