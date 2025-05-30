package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jasonstubblefield/ingredients"
	log "github.com/schollz/logger"
)

func main() {
	log.SetLevel("error")
	if len(os.Args) < 2 {
		log.Error("usage: ingredients [file/url] or ingredients -stdin [name]")
		os.Exit(1)
	}

	var r *ingredients.Recipe
	var err error
	var origin string
	
	type Result struct {
		Ingredients []ingredients.Ingredient `json:"ingredients"`
		Origin      string                   `json:"origin"`
	}

	// Check if reading from stdin
	if os.Args[1] == "-stdin" || os.Args[1] == "--stdin" {
		if len(os.Args) < 3 {
			log.Error("usage: ingredients -stdin [name]")
			os.Exit(1)
		}
		origin = os.Args[2]
		
		// Read HTML from stdin
		htmlBytes, err := ioutil.ReadAll(os.Stdin)
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
		origin = os.Args[1]
		
		h := md5.New()
		io.WriteString(h, origin)
		fname := fmt.Sprintf("%x.json", h.Sum(nil))
		if _, err := os.Stat(fname); err == nil {
			return
		}
		
		r, err = ingredients.NewFromFile(origin)
		if err != nil {
			r, err = ingredients.NewFromURL(origin)
			if err != nil {
				log.Error("usage: ingredients [file/url] or ingredients -stdin [name]")
				os.Exit(1)
			}
		}
	}
	ing := r.IngredientList()
	var re Result
	re.Ingredients = ing.Ingredients
	re.Origin = origin
	
	// Only cache non-stdin inputs
	if os.Args[1] != "-stdin" && os.Args[1] != "--stdin" {
		h := md5.New()
		io.WriteString(h, origin)
		fname := fmt.Sprintf("%x.json", h.Sum(nil))
		
		if len(ing.Ingredients) > 0 {
			b, _ := json.MarshalIndent(re, "", "    ")
			ioutil.WriteFile(fname, b, 0644)
			fmt.Printf("wrote '%s'\n", fname)
		} else {
			fmt.Printf("no ingredients for '%s'\n", origin)
		}
	} else {
		// For stdin, just output the JSON to stdout
		if len(ing.Ingredients) > 0 {
			b, _ := json.MarshalIndent(re, "", "    ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("no ingredients found\n")
		}
	}
}
