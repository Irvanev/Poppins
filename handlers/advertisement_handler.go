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

// Create создаёт новое объявление с загрузкой фотографий.
// @Summary      Создать объявление
// @Description  Создаёт новое объявление пользователя и загружает файлы фото в объектное хранилище.
// @Tags         ads
// @Accept       multipart/form-data
// @Param        user_id      formData  int     true  "ID пользователя"
// @Param        title        formData  string  true  "Заголовок объявления"
// @Param        description  formData  string  true  "Описание объявления"
// @Param        price        formData  int     true  "Цена объявления"
// @Param        address      formData  string  true  "Адрес размещения объявления"
// @Param        photos       formData  []file  true  "Файлы фотографий объявления" collectionFormat(multi)
// @Success      201          {object}  domain.Advertisement
// @Failure      400          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Router       /ads [post]
func (h *AdHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "cannot parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	title := r.FormValue("title")
	description := r.FormValue("description")
	price, _ := strconv.ParseInt(r.FormValue("price"), 10, 64)
	address := r.FormValue("address")

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
	}
	if err := h.Repo.Create(ad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ad)
}

// Get возвращает объявление по его ID.
// @Summary      Получить объявление
// @Description  Возвращает детали объявления по переданному идентификатору.
// @Tags         ads
// @Param        id   path      int  true  "ID объявления"
// @Success      200  {object}  domain.Advertisement
// @Failure      404  {object}  map[string]string
// @Router       /ads/{id} [get]
func (h *AdHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	ad, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ad)
}

// Search обрабатывает поиск объявлений.
// @Summary      Поиск объявлений
// @Description  Ищет объявления по ключевому слову в заголовке и/или по максимальной цене.
// @Tags         ads
// @Param        search     query     string  false  "Ключевое слово для поиска"
// @Param        max_price  query     int     false  "Максимальная цена"
// @Success      200        {array}   domain.Advertisement
// @Failure      400        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /ads [get]
func (h *AdHandler) Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

// Update изменяет существующее объявление.
// @Summary      Обновить объявление
// @Description  Обновляет поля объявления по его ID.
// @Tags         ads
// @Accept       json
// @Param        id   path      int                    true  "ID объявления"
// @Param        ad   body      domain.Advertisement   true  "Объект объявления"
// @Success      200  {object}  domain.Advertisement
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /ads/{id} [put]
func (h *AdHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

// Delete удаляет объявление по ID.
// @Summary      Удалить объявление
// @Description  Удаляет объявление из БД по его идентификатору.
// @Tags         ads
// @Param        id   path      int  true  "ID объявления"
// @Success      204  {string}  string  "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /ads/{id} [delete]
func (h *AdHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err := h.Repo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Archive архивирует объявление (ставит флаг archived = true).
// @Summary      Архивировать объявление
// @Description  Помечает объявление как архивное (archived = true).
// @Tags         ads
// @Param        id   path      int  true  "ID объявления"
// @Success      204  {string}  string  "No Content"
// @Failure      500  {object}  map[string]string
// @Router       /ads/{id}/archive [post]
func (h *AdHandler) Archive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err := h.Repo.Archive(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
