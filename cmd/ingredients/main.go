package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jasonstubblefield/ingredients"
	log "github.com/schollz/logger"
)

// getCacheDir returns the cache directory path, creating it if necessary
func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(homeDir, ".cache", "ingredients")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}
	return cacheDir, nil
}

func main() {
	log.SetLevel("error")

	// Reorder args to put flags first (Go's flag package requires this)
	// This allows users to put flags anywhere: "ingredients url -o file" or "ingredients -o file url"
	var reorderedArgs []string
	var positionalArgs []string
	skipNext := false
	for i := 1; i < len(os.Args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		arg := os.Args[i]
		if arg == "-o" || arg == "--o" {
			if i+1 < len(os.Args) {
				reorderedArgs = append(reorderedArgs, arg, os.Args[i+1])
				skipNext = true
			}
		} else if !skipNext {
			positionalArgs = append(positionalArgs, arg)
		}
	}
	os.Args = append([]string{os.Args[0]}, append(reorderedArgs, positionalArgs...)...)

	// Define flags
	outputFile := flag.String("o", "", "save output to file")
	flag.Parse()

	// Get non-flag arguments
	args := flag.Args()
	if len(args) < 1 {
		log.Error("usage: ingredients [file/url] [-o output.json]")
		log.Error("       ingredients -stdin [name] [-o output.json]")
		os.Exit(1)
	}

	var r *ingredients.Recipe
	var origin string

	type Result struct {
		Ingredients []ingredients.Ingredient `json:"ingredients"`
		Origin      string                   `json:"origin"`
	}

	// Check if reading from stdin
	if args[0] == "-stdin" || args[0] == "--stdin" {
		if len(args) < 2 {
			log.Error("usage: ingredients -stdin [name] [-o output.json]")
			os.Exit(1)
		}
		origin = args[1]

		// Read HTML from stdin
		htmlBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Errorf("failed to read from stdin: %v", err)
			os.Exit(1)
		}
		
		r, err = ingredients.NewFromHTML(origin, string(htmlBytes))
		if err != nil {
			log.Errorf("failed to parse HTML: %v", err)
			os.Exit(1)
		}
	} else {
		origin = args[0]

		// Get cache directory
		cacheDir, err := getCacheDir()
		if err != nil {
			log.Errorf("failed to get cache directory: %v", err)
			os.Exit(1)
		}

		// Check cache first
		h := md5.New()
		io.WriteString(h, origin)
		cachePath := filepath.Join(cacheDir, fmt.Sprintf("%x.json", h.Sum(nil)))

		// Try to load from cache
		if cachedData, err := os.ReadFile(cachePath); err == nil {
			var cached Result
			if json.Unmarshal(cachedData, &cached) == nil {
				// Output cached result to stdout
				b, _ := json.MarshalIndent(cached, "", "    ")
				fmt.Println(string(b))

				// Optionally save to file if -o flag provided
				if *outputFile != "" {
					if err := os.WriteFile(*outputFile, b, 0644); err != nil {
						log.Errorf("failed to write output file: %v", err)
					}
				}
				return
			}
		}

		// Not in cache, fetch it
		r, err = ingredients.NewFromFile(origin)
		if err != nil {
			r, err = ingredients.NewFromURL(origin)
			if err != nil {
				log.Error("usage: ingredients [file/url] [-o output.json]")
				os.Exit(1)
			}
		}
	}
	ing := r.IngredientList()
	var re Result
	re.Ingredients = ing.Ingredients
	re.Origin = origin

	// Marshal to JSON
	b, _ := json.MarshalIndent(re, "", "    ")

	// Save to cache for non-stdin inputs
	if args[0] != "-stdin" && args[0] != "--stdin" {
		cacheDir, err := getCacheDir()
		if err == nil {
			h := md5.New()
			io.WriteString(h, origin)
			cachePath := filepath.Join(cacheDir, fmt.Sprintf("%x.json", h.Sum(nil)))
			os.WriteFile(cachePath, b, 0644)
		}
	}

	// Always output to stdout
	if len(ing.Ingredients) > 0 {
		fmt.Println(string(b))
	} else {
		log.Error("no ingredients found")
		os.Exit(1)
	}

	// Optionally save to file if -o flag provided
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, b, 0644); err != nil {
			log.Errorf("failed to write output file: %v", err)
			os.Exit(1)
		}
	}
}
