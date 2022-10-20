package main

import (
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
	"time"

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

func MustAtoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	return i
}

func MustAtoiOrDefault(val string, orElse int) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		return orElse
	}
	return i
}

func handleUploadAndGenerateCalendar(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	//image upload folder
	err := os.MkdirAll(os.Getenv("CALENDAGO_WORK_DIR"), os.ModePerm)
	if err != nil {
		http.Error(w, "Folder not found", http.StatusInternalServerError)
		return
	}

	form := r.MultipartForm

	year, err := strconv.Atoi(r.FormValue("Year"))
	if r.FormValue("Year") == "" || err != nil {
		http.Error(w, "Please specify Year", http.StatusBadRequest)
		return
	}

	fontName := "arial"
	if r.FormValue("HeaderFont") != "" {
		fontName = r.FormValue("HeaderFont")
	}
	// todo: find out if fiber has this binding for multipart forms
	var settings = generator.Settings{
		Year:              year,
		Width:             MustAtoiOrDefault(r.FormValue("Width"), 1404),
		Height:            MustAtoiOrDefault(r.FormValue("Height"), 1872),
		MarginLeft:        MustAtoiOrDefault(r.FormValue("MarginLeft"), 100),
		MarginRight:       MustAtoiOrDefault(r.FormValue("MarginRight"), 10),
		MarginTop:         MustAtoiOrDefault(r.FormValue("MarginTop"), 5),
		MarginBottom:      MustAtoiOrDefault(r.FormValue("MarginBottom"), 200),
		HeaderFont:        fontName,
		HeaderFontSize:    MustAtoiOrDefault(r.FormValue("HeaderFontSize"), 25),
		StartOfTheWeek:    time.Monday,    // TODO r.FormValue("StartOfTheWeek"),
		CalendarWeek:      generator.None, //r.FormValue("CalendarWeek"),
		CalendarWeekColor: 0.0,
	}

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
		maxFileSize = 3_000_000
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
	//http.Redirect(w, r, "/?Succes=true", http.StatusSeeOther)
}

func main() {
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Printf("Please configure via a .env file")
		return
	}

	http.HandleFunc("/calendar", handleUploadAndGenerateCalendar)

	log.Println("Server started")

	address := os.Getenv("CALENDAGO_ADDRESS")
	if address == "" {
		address = "localhost:8001"
	}

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}
