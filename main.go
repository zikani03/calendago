package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/joho/godotenv"
)

var (
	templateFile = template.Must(template.ParseFiles("static/index.html"))
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	templateFile.ExecuteTemplate(w, "index.html", nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	//image upload folder
	err := os.MkdirAll(os.Getenv("CALENDAGO_WORK_DIR"), os.ModePerm)
	if err != nil {
		http.Error(w, "Folder not found", http.StatusInternalServerError)
		return
	}

	form := r.MultipartForm

	// handle multiple images in the "image" field of the request
	imageFiles, ok := form.File["image"]
	if !ok {
		fmt.Println("No images detected in the request")
		http.Error(w, fmt.Errorf("no images detected").Error(), http.StatusInternalServerError)
		return
	}

	maxFileSize, err := strconv.Atoi(os.Getenv("CALENDAGO_MAX_FILE_SIZE"))
	if err != nil {
		// default if env var is not set
		maxFileSize = 30000
	}

	for _, fileHeader := range imageFiles {
		fmt.Println(fileHeader.Filename)
		if fileHeader.Size > int64(maxFileSize) {
			http.Error(w, "file too big", http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		defer file.Close()

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filename := path.Base(fileHeader.Filename)

		filename = filepath.Join(os.Getenv("CALENDAGO_WORK_DIR"), filename)

		dest, err := os.Create(filename)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer dest.Close()

		if _, err = io.Copy(dest, file); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/?Succes=true", http.StatusSeeOther)
}

func main() {
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Printf("Did not Load env")
	}
	// os.setenv()
	// fmt.Println("Path:", os.Getenv("CALENDAGO_IMAGE_DIR"))
	// fmt.Printf("PATH: %s", os.Getenv("UP_IMAGE_PATH"))

	http.HandleFunc("/", uploadFile)
	http.HandleFunc("/upload", handleUpload)

	log.Println("Server started")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}

}
