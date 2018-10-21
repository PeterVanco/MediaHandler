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
	"io/ioutil"
	"log"
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
	path        string

	// globals
	store *duplo.Store
	bar   *mpb.Bar
)

func init() {

	findDuplicatesCmd = &cobra.Command{
		Use:   "find-duplicates",
		Short: "finds duplicates in in defined path",
		Run:   findDuplicates,
	}

	RootCmd.AddCommand(findDuplicatesCmd)
	findDuplicatesCmd.Flags().IntVar(&sensitivity, "sensitivity", 0, "the sensitivity threshold (the lower, the better the match (can be negative))")
	findDuplicatesCmd.Flags().StringVar(&path, "path", ".", "the path to search the images")
	findDuplicatesCmd.Flags().Int("uidlen", 8, "length of duplicate UID")
}

func findDuplicates(cmd *cobra.Command, args []string) {
	bar = initBar()
	store = duplo.New()

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	Logger.Printf("found %d files\n", len(files))
	bar.SetTotal(int64(len(files)), false)

	distanceMap := make(clustering.DistanceMap)

	for _, f := range files {
		handleFile(f, distanceMap)
		bar.Increment()
	}

	Logger.Printf("calculating clusters from %d hashes\n", len(store.IDs()))
	calculateClusters(clustering.NewDistanceMapClusterSet(distanceMap))
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

func calculateClusters(clusterSet clustering.ClusterSet) {
	clustering.Cluster(clusterSet, clustering.Threshold(float64(sensitivity)), clustering.CompleteLinkage())
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

	if _, extensionSupported := constants.PhotoExtensions[ext]; !extensionSupported {
		return
	}

	filename := filepath.Join(path, f.Name())
	file, err := os.Open(filename)
	if err != nil {
		Logger.Printf("%s: %v\n", filename, err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		Logger.Printf("%s: %v\n", filename, err)
		return
	}
	handleImage(filename, img, f, distanceMap)
}

func handleImage(filename string, img image.Image, f os.FileInfo, distanceMap clustering.DistanceMap) {
	// Add image "img" to the store.
	Logger.Printf("hashing %s\n", f.Name())
	hash, _ := duplo.CreateHash(img)
	matches := store.Query(hash)
	store.Add(f, hash)
	distanceItem := make(map[clustering.ClusterItem]float64)
	if len(matches) > 0 {
		sort.Sort(matches)
		Logger.Printf("adding %d matches to cluster set\n", len(matches))

		for _, match := range matches {
			fileInfo := match.ID.(os.FileInfo)
			// Logger.Printf("    %s score with %s (%d)\n", filename, fileInfo.Name(), int(match.Score))
			distanceItem[fileInfo.Name()] = match.Score
		}

		//match := matches[0]
		//fi := match.ID.(os.FileInfo)
		//// Logger.Printf("%s closest score with %s (%d)\n", filename, fi.Name(), int(match.Score))
		//if int(match.Score) <= *sensitivity {
		//	Logger.Printf("%s matches: %s\n", filename, fi.Name())
		//}
	} else {
		Logger.Println("file has no matches ", filename)
	}
	distanceMap[f.Name()] = distanceItem
}

func moveFile(filename string, uid int64) {
	filenameSuffix := filename
	if match := constants.DuplicateRegExp.FindString(filename); match != "" {
		filenameSuffix = strings.TrimPrefix(filename, match)
	}
	newFilename := fmt.Sprintf("%s_%d_%s", "DUP", uid, filenameSuffix)
	if dryRun, _ := findDuplicatesCmd.Flags().GetBool("dryrun"); dryRun {
		Logger.Printf("    would move file %s to %s\n", filename, newFilename)
	} else {
		if err := os.Rename(filepath.Join(path, filename), filepath.Join(path, newFilename)); err != nil {
			// Logger.Printf("error moving file %s to %s\n", filename, newFilename)
		}
	}
}
