package pdf

import (
	"bytes"
	"fmt"
	"go-typst-pdf/storage"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/template"
	"time"
)

var (
	TemplateCache = make(map[string]*template.Template)
	CacheMutex    sync.RWMutex
)

// Load all templates into in-memory cache at startup
func InitTemplateCache() {
	templatesPath := "templates"
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		panic("Failed to read templates directory: " + err.Error())
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(templatesPath, entry.Name()))
		if err != nil {
			fmt.Println("Failed to read file:", entry.Name(), err)
			continue
		}
		tmpl, err := template.New(entry.Name()).Parse(string(content))
		if err != nil {
			fmt.Println("Failed to parse template:", entry.Name(), err)
			continue
		}
		CacheMutex.Lock()
		TemplateCache[entry.Name()] = tmpl
		CacheMutex.Unlock()
		fmt.Println("Cached template:", entry.Name())
	}
}

// Try to get template from memory cache, fallback to disk read
func getCachedTemplate(name string) (*template.Template, error) {
	CacheMutex.RLock()
	tmpl, ok := TemplateCache[name]
	CacheMutex.RUnlock()
	if ok {
		return tmpl, nil
	}
	// Fallback: read from disk and cache
	path := filepath.Join("templates", name)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tmpl, err = template.New(name).Parse(string(content))
	if err != nil {
		return nil, err
	}
	CacheMutex.Lock()
	TemplateCache[name] = tmpl
	CacheMutex.Unlock()
	return tmpl, nil
}

// Job represents a PDF generation job
type Job struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

func GeneratePDFsInParallel(jobs []Job) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	for _, job := range jobs {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire a slot
			defer func() { <-sem }() // Release the slot

			_, err := GenerateAndUpload(job.Template, job.Data)
			if err != nil {
				fmt.Printf("Error generating PDF for template %s: %v", job.Template, err)
			}
		}(job)
	}

	wg.Wait()
}

func GenerateAndUpload(templateName string, data map[string]interface{}) (string, error) {
	tmpl, err := getCachedTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	// Save rendered Typst source to a temp file
	inFile := fmt.Sprintf("/dev/shm/invoice_%d.typ", time.Now().UnixNano())
	outFile := fmt.Sprintf("/dev/shm/invoice_%d.pdf", time.Now().UnixNano())

	// Ensure the output directory exists
	outputDir := "output/pdf"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(inFile, rendered.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp .typ file: %w", err)
	}

	defer os.Remove(inFile) // Clean up after compile

	cmd := exec.Command("typst", "compile", inFile, outFile)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout // Optional: log to console

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("typst compile failed: %v, stderr: %s", err, stderr.String())
	}

	// Upload PDF and return URL
	url, err := storage.UploadPDF(outFile)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return url, nil
}
