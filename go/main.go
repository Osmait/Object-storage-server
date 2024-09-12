package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Bucket struct {
	Name string `json:name`
}

func main() {
	mux := http.NewServeMux()
	addr := ":8080"
	mux.HandleFunc("POST /uploaded", uploadFile)
	mux.HandleFunc("POST /create", createBucket)
	mux.HandleFunc("/list", getFiles)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello word")
	})
	fmt.Printf("server listenen to %s\n", addr)
	http.ListenAndServe(":8080", mux)
}

func getFiles(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {

		http.Error(w, "name is require", 400)
		return
	}
	fmt.Println(name)
	objects := []string{}
	err := filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		objects = append(objects, path)
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}
	fmt.Fprintf(w, "%s", objects)
}

func createBucket(w http.ResponseWriter, r *http.Request) {
	var bucket Bucket
	err := json.NewDecoder(r.Body).Decode(&bucket)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	os.Mkdir(bucket.Name, 0777)
	fmt.Fprint(w, bucket.Name)
}

// Función para manejar la subida del archivo
func uploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100 << 20) // Limite de tamaño de 10 MB
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create(fmt.Sprintf("valor1/%s", handler.Filename))
	if err != nil {
		http.Error(w, "Error creating file on server", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}
