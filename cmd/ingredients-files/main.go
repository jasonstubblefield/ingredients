package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jasonstubblefield/ingredients"
	log "github.com/schollz/logger"
	"github.com/schollz/progressbar/v3"
)

func main() {
	log.SetLevel("error")
	fnames := []string{}
	err := filepath.Walk(os.Args[1],
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				fnames = append(fnames, path)
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var numJobs = len(fnames)

	type job struct {
		pathToFile string
	}

	type result struct {
		err error
	}

	jobs := make(chan job, numJobs)
	results := make(chan result, numJobs)
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(jobs <-chan job, results chan<- result) {
			for j := range jobs {
				// step 3: specify the work for the worker
				var r result
				processFile(os.Args[1], j.pathToFile)
				results <- r
			}
		}(jobs, results)
	}

	// step 4: send out jobs
	for i := 0; i < numJobs; i++ {
		jobs <- job{fnames[i]}
	}
	close(jobs)

	// step 5: do something with results
	b := progressbar.Default(int64(numJobs))
	for i := 0; i < numJobs; i++ {
		b.Add(1)
		r := <-results
		if r.err != nil {
			// do something with error
		}
	}
}

func processFile(folderName, pathToFile string) {
	h := md5.New()
	io.WriteString(h, pathToFile)
	fname := fmt.Sprintf("%x.json", h.Sum(nil))
	finalPath := path.Join(folderName, fname)

	// Quick check if file already exists (optimization to skip processing)
	if _, err := os.Stat(finalPath); err == nil {
		return
	}

	var r *ingredients.Recipe
	var err error
	type Result struct {
		Ingredients []ingredients.Ingredient `json:"ingredients"`
		Origin      string                   `json:"origin"`
	}
	r, err = ingredients.NewFromFile(pathToFile)
	if err != nil {
		log.Errorf("%s: %s", pathToFile, err)
		return
	}
	ing := r.IngredientList()
	var re Result
	re.Ingredients = ing.Ingredients
	re.Origin = strings.TrimPrefix(pathToFile, folderName)
	re.Origin = strings.TrimPrefix(re.Origin, "/")
	if len(ing.Ingredients) >= 2 {
		b, _ := json.MarshalIndent(re, "", "    ")

		// Use atomic file creation: write to temp file then rename
		// This prevents race conditions where multiple goroutines process the same file
		tempPath := finalPath + ".tmp." + fmt.Sprintf("%d", os.Getpid())
		if err := os.WriteFile(tempPath, b, 0644); err != nil {
			log.Errorf("failed to write temp file %s: %s", tempPath, err)
			return
		}

		// Atomic rename - if this fails because file exists, another goroutine won the race
		if err := os.Rename(tempPath, finalPath); err != nil {
			// File might already exist from another goroutine - clean up temp file
			os.Remove(tempPath)
			// Only log if it's not a "file exists" error
			if !os.IsExist(err) {
				log.Errorf("failed to rename %s to %s: %s", tempPath, finalPath, err)
			}
			return
		}

		log.Debugf("wrote '%s'\n", pathToFile)
	} else {
		log.Debugf("insufficient ingredients for '%s': %d (minimum 2 required)\n", pathToFile, len(ing.Ingredients))
	}
}
