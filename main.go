package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/signintech/gopdf"
	"github.com/zikani03/calendago/generator"
)

var (
	templateFile = template.Must(template.ParseFiles("static/index.html"))
)

func getAllImages(year int) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(os.Getenv("CALENDAGO_WORK_DIR"))
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		if strings.Contains(file.Name(), ".png") && strings.Contains(file.Name(), strconv.Itoa(year)+"_") {
			files = append(files, filepath.Join(os.Getenv("CALENDAGO_WORK_DIR"), file.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

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

func generateCalendar(w http.ResponseWriter, r *http.Request) {
	// load layout setting and parse them into an object
	var settings generator.Settings
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, &settings)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// generate the calendar pages
	generator.Generate(settings, os.Getenv("CALENDAGO_WORK_DIR"))

	images, err := getAllImages(settings.Year)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	for i := 0; i < len(images); i++ {
		pdf.AddPage()
		pdf.Image(images[i], 0, 0, gopdf.PageSizeA4)
	}

	pdfFilename := filepath.Join(os.Getenv("CALENDAGO_WORK_DIR"), strconv.Itoa(settings.Year)+".pdf")
	// Save to disk as backup
	pdf.WritePdf(pdfFilename)
	pdfData, err := ioutil.ReadFile(pdfFilename)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Disposition", "filename=calendar.pdf")
	w.Header().Add("Content-Type", "application/pdf")
	w.Write(pdfData)
}

func main() {
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Printf("Please configure via a .env file")
		return
	}

	http.HandleFunc("/", uploadFile)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/calendar", generateCalendar)

	log.Println("Server started")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}

}
