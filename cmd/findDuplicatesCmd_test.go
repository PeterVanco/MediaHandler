package cmd

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func Test_findDuplicates(t *testing.T) {
	type args struct {
		path   string
		dryRun bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{
			path:   "test",
			dryRun: true,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Chdir("/home/vancop/repo/MediaHandler")
			fmt.Printf("%v\n", err)
			a, err := os.Executable()
			fmt.Printf("%s, %v\n", a, err)
			findDuplicates(tt.args.path, tt.args.dryRun)
		})
	}
}

func Test_findImages(t *testing.T) {
	testPath := "/home/vancop/repo/MediaHandler/test"
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Find images",
			args: args{
				path: testPath,
			},
			want: []string{
				testPath + "/mixed/red2.jpg",
				testPath + "/reds/red1.jpg",
				testPath + "/reds/red2.jpg",
				testPath + "/yellow/yellow.jpg",
				testPath + "/greens/green2.jpg",
				testPath + "/mixed/green1.jpg",
				testPath + "/mixed/red1.jpg",
				testPath + "/greens/green1.jpg",
				testPath + "/mixed/green2.jpg",
				testPath + "/mixed/yellow.jpg",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findImages(tt.args.path)
			containsAllAndOnly(t, result, tt.want)
		})
	}
}

func containsAllAndOnly(t *testing.T, got, expected []string) {
	containsAll(t, expected, got)
	containsAll(t, got, expected)
}

func containsAll(t *testing.T, expected, got []string) {
	sliceMap := make(map[string]bool)
	for _, gotItem := range got {
		sliceMap[gotItem] = true
	}

	for _, expectedItem := range expected {
		if found, _ := sliceMap[expectedItem]; !found {
			t.Errorf("expected item %v not present in result", expectedItem)
			return
		}
	}
}

func containsAll2(t *testing.T, expected, got []string) {
	for _, expectedItem := range expected {
		log.Printf("expected: %v\n", expectedItem)
		var found bool
		for _, gotItem := range got {
			log.Printf("got: %v\n", gotItem)
			if reflect.DeepEqual(expectedItem, gotItem) {
				found = true
			}
		}
		if !found {
			t.Errorf("item %v not present in result", expectedItem)
			return
		}
	}
}
