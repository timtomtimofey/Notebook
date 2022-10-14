package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ImageHandler struct {
	path       string
	fileServer http.Handler
}

func NewImageHandler(path string) (*ImageHandler, error) {
	err := os.Mkdir(path, 0750)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	res := &ImageHandler{
		path:       path,
		fileServer: http.FileServer(http.Dir(path)),
	}
	return res, nil
}

func (i *ImageHandler) Get(w http.ResponseWriter, r *http.Request) {
	last := len(r.URL.Path) - 1
	if last >= 0 && r.URL.Path[last] == '/' {
		r.URL.Path = r.URL.Path[:last]
	}
	// log.Print(r.URL)
	i.fileServer.ServeHTTP(w, r)
}

func (i *ImageHandler) Add(w http.ResponseWriter, r *http.Request) {
	data, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Cant parse image form: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer data.Close()

	id := uuid.NewString()
	path := filepath.Join(i.path, id)
	file, err := os.Create(path)
	if err != nil {
		http.Error(w, "Cant create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	hash := md5.New()
	multiWriter := io.MultiWriter(file, hash)
	if _, err := io.Copy(multiWriter, data); err != nil {
		http.Error(w, "Cant write image: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sum := hash.Sum([]byte{})
	sumStr := hex.EncodeToString(sum)
	hashPath := filepath.Join(i.path, sumStr)
	if err := os.Rename(path, hashPath); err != nil {
		http.Error(w, "Cant rename file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{ \"image_id\": \"" + sumStr + "\" }"))
}

func (i *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || id == "" {
		http.Error(w, "error parsing {id}", http.StatusInternalServerError)
		return
	}

	path := filepath.Join(i.path, id)
	if err := os.Remove(path); err != nil {
		http.Error(w, "cant delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
