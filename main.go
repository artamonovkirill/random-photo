package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MaestroError/go-libheif"
)

type file struct {
	path     string
	modified time.Time
}

type pointer struct {
	Current, Previous, Next int
	Path, Origin, Ext       string
}

const port = ":8080"

var filesRoot string
var fs []file

func rootHandler(w http.ResponseWriter, r *http.Request) {
	current := rand.Intn(len(fs))
	http.Redirect(w, r, fmt.Sprintf("/photos/%d", current), http.StatusSeeOther)
}

func photoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/photos/"))
	if err != nil {
		fmt.Println(err)
		return
	}
	path := fs[id].path
	origin := strings.ReplaceAll(fs[id].path, filesRoot, "")
	ext := filepath.Ext(path)

	switch ext {
	case ".HEIC":
		err := libheif.HeifToJpeg(path, "./tmp/current.jpeg", 100)
		if err != nil {
			fmt.Println(err)
			return
		}
		path = "/tmp/current.jpeg"
	default:
		path = fmt.Sprintf("http://localhost%s/static%s", port, origin)
	}

	t, _ := template.ParseFiles("main.html")
	t.Execute(w, pointer{
		Current:  id,
		Previous: id - 1,
		Next:     id + 1,
		Path:     path,
		Origin:   origin,
		Ext:      ext,
	})
}

func main() {
	filesRoot = os.Getenv("PHOTO_ROOT")
	if filesRoot == "" {
		fmt.Println("PHOTO_ROOT environment variable is not set")
		os.Exit(1)
	}

    fmt.Println("Scanning the library")
	fs = files(filesRoot)
	fmt.Printf("Discovered %d files\n", len(fs)
	sort.Slice(fs, func(i, j int) bool {
		return fs[i].modified.Before(fs[j].modified)
	})

	os.Mkdir("./tmp", os.ModePerm)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/photos/", photoHandler)

	photos := http.FileServer(http.Dir(filesRoot))
	http.Handle("/static/", http.StripPrefix("/static/", photos))
	assets := http.FileServer(http.Dir("."))
	http.Handle("/tmp/", assets)
	http.Handle("/favicon.ico", assets)

	log.Fatal(http.ListenAndServe(port, nil))
}

func files(path string) []file {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var result []file
	for _, entry := range entries {
		entryPath := path + "/" + entry.Name()
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			result = append(result, file{
				path:     entryPath,
				modified: info.ModTime(),
			})
		} else {
			result = append(result, files(entryPath)...)
		}
	}
	return result
}
