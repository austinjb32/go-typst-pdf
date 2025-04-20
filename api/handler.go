package api

import (
	"encoding/json"
	"go-typst-pdf/queue"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type PDFRequest struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	var req PDFRequest
	json.NewDecoder(r.Body).Decode(&req)

	job := queue.Job{
		Template: req.Template,
		Data:     req.Data,
	}
	queue.AddJobToQueue(job)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Job added to queue"))
}

type TemplateRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func ListTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	templatesPath := "templates"
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		http.Error(w, "Failed to read templates directory", http.StatusInternalServerError)
		return
	}

	var templates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func UploadFormHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/insert_template.html")
}

func UploadTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // Limit upload size to 10MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file
	file, handler, err := r.FormFile("templateFile")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate the file extension
	if filepath.Ext(handler.Filename) != ".typ" {
		http.Error(w, "Only .typ files are allowed", http.StatusBadRequest)
		return
	}

	// Save the file to the `pdf/templates` directory
	templatePath := filepath.Join("templates", handler.Filename)
	dst, err := os.Create(templatePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Template uploaded successfully"))
}
