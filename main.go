package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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

	files, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Bad Requeat", http.StatusBadRequest)
		return
	}
	//upload multiple images
	for _, fileHeader := range files {
		if fileHeader.Size > os.Getenv("CALENDAGO_MAX_FILE_SIZE") {
			http.Error(w, "file too big", http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		defer files.Close()

		, err := files.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		

		//image upload folder
		err = os.MkdirAll(os.Getenv("CALENDAGO_WORK_DIR"), os.ModePerm)
		if err != nil {
			http.Error(w, "Folder not found", http.StatusInternalServerError)
		}

		filename := path.Base(fileHeader.Filename)
		dest, err := os.Create(filename)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer dest.Close()

		if _, err = io.Copy(dest, files); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/?Succes=true", http.StatusSeeOther)
		}	
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

	log.Println("Server started")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}

	
}
