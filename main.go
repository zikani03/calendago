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

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Bad Requeat", http.StatusBadRequest)
		return
	}
	defer file.Close()

	//image upload folder
	err = os.MkdirAll(os.Getenv("UP_IMAGE_PATH "), os.ModePerm)
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

	if _, err = io.Copy(dest, file); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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

	log.Println("Server started")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}

	
}
