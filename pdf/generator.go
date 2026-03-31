package pdf

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"go-typst-pdf/storage"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/template"
)

const templatesPath = "pdf/templates"

var (
	TemplateCache = make(map[string]*template.Template)
	CacheMutex    sync.RWMutex
)

func mustRandBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

// Load all templates into in-memory cache at startup
func InitTemplateCache() {
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

// InvalidateTemplateCache removes a template from the in-memory cache
func InvalidateTemplateCache(name string) {
	CacheMutex.Lock()
	delete(TemplateCache, name)
	CacheMutex.Unlock()
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
	path := filepath.Join(templatesPath, name)
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

func GeneratePDFsInParallel(jobs []Job) []error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	sem := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	for _, job := range jobs {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire a slot
			defer func() { <-sem }() // Release the slot

			_, err := GenerateAndUpload(job.Template, job.Data)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("template %s: %w", job.Template, err))
				mu.Unlock()
			}
		}(job)
	}

	wg.Wait()
	return errs
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

	// Use a single unique ID for both temp files to prevent collision
	id := fmt.Sprintf("%x", mustRandBytes(16))
	inFile := filepath.Join(os.TempDir(), fmt.Sprintf("typst_%s.typ", id))
	outFile := filepath.Join(os.TempDir(), fmt.Sprintf("typst_%s.pdf", id))

	if err := os.WriteFile(inFile, rendered.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp .typ file: %w", err)
	}
	defer os.Remove(inFile) // Clean up .typ source

	cmd := exec.Command("typst", "compile", inFile, outFile)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.Remove(outFile) // Clean up partial PDF on failure
		return "", fmt.Errorf("typst compile failed: %v, stderr: %s", err, stderr.String())
	}

	// Upload PDF and return URL
	url, err := storage.UploadPDF(outFile)
	os.Remove(outFile) // Always clean up PDF temp file
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return url, nil
}
