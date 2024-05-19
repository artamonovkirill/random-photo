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
	"syscall"
	"time"

	"github.com/MaestroError/go-libheif"
	"github.com/google/uuid"
)

type file struct {
	path    string
	created time.Time
}

type pointer struct {
	Current, Previous, Next             int
	Path, Origin, Ext, Session, Created string
}

const DATE = "02 January 2006"

var filesRoot string
var fs []file

func rootHandler(w http.ResponseWriter, r *http.Request) {
	current := rand.Intn(len(fs))
	session := r.URL.Query().Get("session")
	if session == "" {
		session = uuid.New().String()
	}
	http.Redirect(w, r, fmt.Sprintf("/photos/%d?session=%s", current, session), http.StatusSeeOther)
}

func photoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/photos/"))
	if err != nil {
		fmt.Println(err)
		return
	}
	session := r.URL.Query().Get("session")
	if session == "" {
		session = uuid.New().String()
	}

	file := fs[id]
	path := file.path
	origin := strings.ReplaceAll(path, filesRoot, "")
	ext := strings.ToUpper(filepath.Ext(path))

	switch ext {
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

	t, _ := template.ParseFiles("main.html")
	t.Execute(w, pointer{
		Current:  id,
		Previous: id - 1,
		Next:     id + 1,
		Path:     path,
		Origin:   origin,
		Ext:      ext,
		Session:  session,
		Created:  file.created.Format(DATE),
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
	fmt.Printf("Discovered %d files\n", len(fs))
	sort.Slice(fs, func(i, j int) bool {
		return fs[i].created.Before(fs[j].created)
	})

	os.Mkdir("./tmp", os.ModePerm)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/photos/", photoHandler)

	files := http.FileServer(http.Dir(filesRoot))
	http.Handle("/static/", http.StripPrefix("/static/", files))

	tmp := http.FileServer(http.Dir("./tmp"))
	http.Handle("/tmp/", http.StripPrefix("/tmp/", tmp))

	log.Fatal(http.ListenAndServe(":8080", nil))
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
			var created time.Time
			switch info.Sys().(type) {
			case *syscall.Stat_t:
				sys := info.Sys().(*syscall.Stat_t)
				created = time.Unix(sys.Birthtimespec.Sec, sys.Birthtimespec.Nsec)
			default:
				fmt.Printf("got: %T\n", info.Sys())
				created = info.ModTime()
			}
			result = append(result, file{
				path:    entryPath,
				created: created,
			})
		} else {
			result = append(result, files(entryPath)...)
		}
	}
	return result
}
