package service

import (
	"fmt"
	"github.com/go-chi/jwtauth"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadSize = 2 * 1024 * 1024 // 2 mb
const uploadPath = "upload"

//  curl -F 'data=@/Users/klim0v/Downloads/IMG_20200320_001309_697.jpg' -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXN0aW5hdGlvbiI6ImF2YXRhciIsInNlcnZpY2UiOiJtaW50ZXIvd2FsbGV0IiwidXNlcl9pZCI6MTIzfQ.RFs4P_BeuDlTOBIVnzQyJOKX0OM_Wl9Bd5ausX2939M" localhost:3333/upload
func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {

	_, claims, _ := jwtauth.FromContext(r.Context())
	userID := fmt.Sprintf("%v", claims["user_id"])
	service := fmt.Sprintf("%v", claims["service"])         // Example: "minter/wallet"
	destination := fmt.Sprintf("%v", claims["destination"]) // Example: "avatar"
	if userID == "" || service == "" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		http.Error(w, "CANT_PARSE_FORM", http.StatusInternalServerError)
		return
	}
	// parse and validate file and post parameters
	file, fileHeader, err := r.FormFile("data")
	if err != nil {
		fmt.Printf("FormFile data err: %v\n", err)
		http.Error(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileSize := fileHeader.Size
	fmt.Printf("File size (bytes): %v\n", fileSize)
	if fileSize > maxUploadSize {
		fmt.Printf("File too big %v\n", fileSize)
		http.Error(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Invalid file err: %v\n", err)
		http.Error(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	dirStorage, availableFileTypes, err := s.repository.StorageInfo(service, destination)
	if err != nil {
		fmt.Printf("Repository StorageInfo err: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// check file type, detect content type only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	splitAvailableFileTypes := strings.Split(strings.ReplaceAll(availableFileTypes, " ", ""), ",")
	var available bool
	for _, availableFileType := range splitAvailableFileTypes {
		if availableFileType == detectedFileType {
			available = true
			break
		}
	}
	if !available {
		http.Error(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
		return
	}

	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil {
		fmt.Printf("Can't read file type err: %v\n", err)
		http.Error(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		return
	}
	newDir := filepath.Join(".", uploadPath, dirStorage, destination)
	//if _, err := os.Stat(newDir); os.IsNotExist(err) {
	err = os.MkdirAll(newDir, os.ModePerm)
	if err != nil {
		fmt.Printf(" %v\n", err)
	}
	//}
	newPath := filepath.Join(newDir, userID+fileEndings[0])
	fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		fmt.Printf("Can't write file err: %v\n", err)
		http.Error(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		fmt.Printf("Can't write file err: %v\n", err)
		http.Error(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("SUCCESS"))
}
