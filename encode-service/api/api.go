package api

import (
	"encode-service/config"
	"encode-service/encode"
	"encode-service/model"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func isFileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func getVideoPath(name string) string {
	videoExtention := []string{".mkv", ".mp4", ".avi"}
	path := ""
	for _, ve := range videoExtention {
		path = filepath.Join(config.VideoFolder, name, name+ve)
		if isFileExist(path) {
			return path
		}
	}
	return ""
}

func encodeHandle(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	movie := vars["movie"]

	path := ""
	if path = getVideoPath(movie); path == "" {
		return model.ClientError{
			Root:     nil,
			Response: fmt.Sprintf(`Cannot find movie "%v"`, movie),
			Status:   http.StatusNotFound}
	}

	err := encode.EncodeVideo(w, path, movie)
	return err
}

func stopEncodeHandle(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	movie := vars["movie"]

	err := encode.StopEncode(w, movie)
	return err
}

func progressHandle(w http.ResponseWriter, r *http.Request) error {
	param := r.URL.Query()["movies"]
	result := make([]encode.EncodeProgress, 0)
	for _, movie := range param {
		path := ""
		if path = getVideoPath(movie); path == "" {
			continue
		}
		progress, err := encode.GetEncodeProgress(path, movie)
		if err != nil {
			return err
		}
		result = append(result, *progress)
	}
	if err := json.NewEncoder(w).Encode(result); err != nil {
		return err
	}
	return nil
}
