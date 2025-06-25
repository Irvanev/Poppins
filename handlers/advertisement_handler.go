package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"poppins/domain"
	"poppins/repository"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

type AdHandler struct {
	Repo        *repository.AdRepo
	MinioClient *minio.Client
	Bucket      string
}

func NewAdHandler(repo *repository.AdRepo, mc *minio.Client, bucket string) *AdHandler {
	return &AdHandler{Repo: repo, MinioClient: mc, Bucket: bucket}
}

func (h *AdHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "cannot parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	title := r.FormValue("title")
	description := r.FormValue("description")
	price, _ := strconv.ParseInt(r.FormValue("price"), 10, 64)
	address := r.FormValue("address")
	archived := r.FormValue("archived") == "true"

	var photoURLs []string
	files := r.MultipartForm.File["photos"]
	ctx := context.Background()
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		objectName := fmt.Sprintf("ads/%d_%d%s", userID, time.Now().UnixNano(), filepath.Ext(fh.Filename))
		info, err := h.MinioClient.PutObject(ctx, h.Bucket, objectName, file, fh.Size, minio.PutObjectOptions{
			ContentType: fh.Header.Get("Content-Type"),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = info
		photoURLs = append(photoURLs, fmt.Sprintf("/%s/%s", h.Bucket, objectName))
	}

	ad := &domain.Advertisement{
		UserID:      userID,
		Title:       title,
		Description: description,
		Price:       price,
		PhotosUrls:  photoURLs,
		Address:     address,
		Archived:    archived,
	}
	if err := h.Repo.Create(ad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ad)
}

func (h *AdHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	ad, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ad)
}

func (h *AdHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	keyword := q.Get("search")
	var maxPrice int64
	if mp := q.Get("max_price"); mp != "" {
		if p, err := strconv.ParseInt(mp, 10, 64); err == nil {
			maxPrice = p
		} else {
			http.Error(w, "invalid max_price", http.StatusBadRequest)
			return
		}
	}
	ads, err := h.Repo.Search(keyword, maxPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ads)
}

func (h *AdHandler) List(w http.ResponseWriter, r *http.Request) {
	ads, err := h.Repo.Search("", 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ads)
}

func (h *AdHandler) Update(w http.ResponseWriter, r *http.Request) {
	var ad domain.Advertisement
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	ad.ID = id
	if err := h.Repo.Update(&ad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ad)
}

func (h *AdHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err := h.Repo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err := h.Repo.Archive(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
