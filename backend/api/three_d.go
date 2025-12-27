package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	Store3DPath = "./store_images/3d_images"
)

// Check3DModel checks if a 3D model exists for a given product image name
// It matches the product image name (without extension) to 3D model files
func Check3DModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method Not Allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	imageName := r.URL.Query().Get("image")
	if imageName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "image parameter is required"})
		return
	}

	// Remove file extension from the image name
	// e.g., "product.jpg" becomes "product"
	imageName = strings.TrimSuffix(imageName, filepath.Ext(imageName))

	// Check for common 3D model file formats
	model3DFormats := []string{".glb", ".gltf", ".obj", ".fbx", ".usdz"}
	has3DModel := false
	var model3DPath string

	// Look for a matching 3D model file
	for _, format := range model3DFormats {
		path := filepath.Join(Store3DPath, imageName+format)
		if _, err := os.Stat(path); err == nil {
			has3DModel = true
			model3DPath = imageName + format
			break
		}
	}

	response := map[string]interface{}{
		"has_3d_model": has3DModel,
		"model_path":   model3DPath,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Get3DModel serves a 3D model file
// Prevents directory traversal attacks by validating the file path
func Get3DModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, `{"error": "model parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Prevent directory traversal attacks
	modelName = filepath.Base(modelName)

	modelPath := filepath.Join(Store3DPath, modelName)

	// Verify the file exists and is in the correct directory
	absPath, err := filepath.Abs(modelPath)
	if err != nil {
		http.Error(w, `{"error": "Invalid path"}`, http.StatusBadRequest)
		return
	}

	absStoreDir, err := filepath.Abs(Store3DPath)
	if err != nil {
		http.Error(w, `{"error": "Invalid path"}`, http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(absPath, absStoreDir) {
		http.Error(w, `{"error": "Access denied"}`, http.StatusForbidden)
		return
	}

	// Set appropriate MIME type based on file extension
	ext := strings.ToLower(filepath.Ext(modelName))
	mimeType := "application/octet-stream"

	switch ext {
	case ".glb":
		mimeType = "model/gltf-binary"
	case ".gltf":
		mimeType = "model/gltf+json"
	case ".obj":
		mimeType = "text/plain"
	case ".mtl":
		mimeType = "text/plain"
	case ".fbx":
		mimeType = "application/octet-stream"
	case ".usdz":
		mimeType = "model/vnd.pixar.usdz+zip"
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, modelPath)
}

// ListProducts3DModels returns a list of products that have 3D models available
func ListProducts3DModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	storeIDStr := r.URL.Query().Get("store_id")
	if storeIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "store_id is required"})
		return
	}

	// Read the 3d_images directory to get available models
	entries, err := os.ReadDir(Store3DPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read 3D models"})
		return
	}

	// Get unique model names (without extensions)
	modelMap := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
			modelMap[nameWithoutExt] = true
		}
	}

	response := map[string]interface{}{
		"available_models": modelMap,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// View3DModel renders the 3D viewer page with the model name injected into the template
func View3DModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, "Model parameter is required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal attacks
	modelName = filepath.Base(modelName)

	// Parse and execute the template
	tmpl, err := template.ParseFiles("./frontend/3dview_template.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"ModelName": modelName,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
