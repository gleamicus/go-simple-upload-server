package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/Sirupsen/logrus"
)

// Server represents a simple-upload server.
type Server struct {
	DocumentRoot string
}

// NewServer creates a new simple-upload server.
func NewServer(documentRoot string) Server {
	return Server{
		DocumentRoot:  documentRoot,
	}
}

func (s Server) handleGet(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`^/snaps/([^/]+)$`)
	if !re.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, fmt.Errorf("\"%s\" is not found", r.URL.Path))
		return
	}
	http.StripPrefix("/snaps/", http.FileServer(http.Dir(s.DocumentRoot))).ServeHTTP(w, r)
}

func (s Server) handlePost(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`^/snaps/([^/]+)$`)
	matches := re.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		logger.WithField("path", r.URL.Path).Info("invalid path")
		w.WriteHeader(http.StatusNotFound)
		writeError(w, fmt.Errorf("\"%s\" is not found", r.URL.Path))
		return
	}
	targetPath := path.Join(s.DocumentRoot, matches[1])
	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.WithError(err).WithField("path", targetPath).Error("failed to open the file")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer file.Close()
	defer r.Body.Close()

	n, err := io.Copy(file, r.Body)
	if err != nil {
		logger.WithError(err).WithField("path", targetPath).Error("failed to write body to the file")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	logger.WithFields(logrus.Fields{
		"path": r.URL.Path,
		"size": n,
	}).Info("file uploaded")
	w.WriteHeader(http.StatusOK)
	writeSuccess(w, r.URL.Path)
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		s.handleGet(w, r)
	case http.MethodPost, http.MethodPut:
		s.handlePost(w, r)
	default:
		w.Header().Add("Allow", "GET,HEAD,POST,PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeError(w, fmt.Errorf("method \"%s\" is not allowed", r.Method))
	}
}
