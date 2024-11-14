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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type file struct {
	path     string
	modified time.Time
	ext      string
}

type pointer struct {
	Previous, Next, Refresh, Path, Origin, Ext, Modified string
}

const DATE = "02 January 2006"

var filesRoot string
var fs []file
var videos []file

func photoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	session := r.URL.Query().Get("session")
	if session == "" {
		session = uuid.New().String()
		http.Redirect(w, r, fmt.Sprintf("/photos/%d?session=%s", id, session), http.StatusSeeOther)
		return
	}

	file := fs[id]
	path := file.path
	origin := strings.ReplaceAll(path, filesRoot, "")

	switch file.ext {
	case ".HEIC":
		output := "/tmp/" + session + ".jpeg"
		err := libheif.HeifToJpeg(path, "."+output, 100)
		if err != nil {
			fmt.Println(err)
			return
		}
		path = output
	default:
		path = "/static" + origin
	}

	t, _ := template.ParseFiles("assets/photos.html")
	t.Execute(w, pointer{
		Previous: fmt.Sprintf("/photos/%d?session=%s", id-1, session),
		Next:     fmt.Sprintf("/photos/%d?session=%s", id+1, session),
		Refresh:  fmt.Sprintf("/photos?session=%s", session),
		Path:     path,
		Origin:   pretty(origin, file.modified),
		Ext:      file.ext,
		Modified: file.modified.Format(DATE),
	})
}

var prefixeFormats = []string{
	"2006 01 January",
	"2006 01 Jan",
	"2006 01",
	"2006",
}

func pretty(origin string, modified time.Time) string {
	var breadcrumbs []string

	var prefixes []string
	for _, f := range prefixeFormats {
		prefixes = append(prefixes, modified.Format(f))
	}

	parts := strings.Split(origin, "/")
	for _, b := range parts[1 : len(parts)-1] {
		for _, p := range prefixes {
			b = strings.TrimPrefix(b, p)
		}
		b = strings.TrimSpace(b)
		if len(b) > 0 {
			breadcrumbs = append(breadcrumbs, b)
		}
	}
	return strings.Join(breadcrumbs, "/")
}

func videoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	file := videos[id]
	path := file.path
	origin := strings.ReplaceAll(path, filesRoot, "")
	path = "/static" + origin

	t, _ := template.ParseFiles("assets/videos.html")
	t.Execute(w, pointer{
		Previous: fmt.Sprintf("/videos/%d", id-1),
		Next:     fmt.Sprintf("/videos/%d", id+1),
		Refresh:  "/videos",
		Path:     path,
		Origin:   pretty(origin, file.modified),
		Ext:      file.ext,
		Modified: file.modified.Format(DATE),
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

	for _, f := range fs {
		if f.ext == ".MP4" || f.ext == ".MOV" {
			videos = append(videos, f)
		}
	}

	fmt.Printf("Discovered %d files, %d videos\n", len(fs), len(videos))
	sort.Slice(fs, func(i, j int) bool {
		return fs[i].modified.Before(fs[j].modified)
	})
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].modified.Before(videos[j].modified)
	})

	os.Mkdir("./tmp", os.ModePerm)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/main.html")
	})
	r.HandleFunc("/photos", func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("/photos/%d", rand.Intn(len(fs)))
		session := r.URL.Query().Get("session")
		if session != "" {
			url += "?session=" + session
		}

		http.Redirect(w, r, url, http.StatusSeeOther)
	})
	r.HandleFunc("/photos/{id}", photoHandler)

	r.HandleFunc("/videos", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/videos/%d", rand.Intn(len(videos))), http.StatusSeeOther)
	})
	r.HandleFunc("/videos/{id}", videoHandler)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(filesRoot))))

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	r.PathPrefix("/tmp/").Handler(http.StripPrefix("/tmp/", http.FileServer(http.Dir("./tmp"))))

	log.Fatal(http.ListenAndServe(":8080", r))
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
				ext:      strings.ToUpper(filepath.Ext(entryPath)),
			})
		} else {
			result = append(result, files(entryPath)...)
		}
	}
	return result
}
