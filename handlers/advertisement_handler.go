package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
// handlers/ad.go
func (h *AdHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ограничим размер формы до 20 МБ
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "cannot parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Читаем telegram_id
	telegramID := r.FormValue("telegram_id")

	title := r.FormValue("title")
	description := r.FormValue("description")
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "invalid price: "+err.Error(), http.StatusBadRequest)
		return
	}
	address := r.FormValue("address")

	// Обрабатываем одно фото
	file, fh, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	objectName := fmt.Sprintf("ads/%s_%d%s",
		telegramID, time.Now().UnixNano(), filepath.Ext(fh.Filename),
	)
	_, err = h.MinioClient.PutObject(
		context.Background(),
		h.Bucket,
		objectName,
		file,
		fh.Size,
		minio.PutObjectOptions{ContentType: fh.Header.Get("Content-Type")},
	)
	if err != nil {
		http.Error(w, "upload error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	photoURL := fmt.Sprintf("/ads/%s", objectName)

	// Собираем объявление
	ad := &domain.Advertisement{
		TelegramID:  telegramID,
		Title:       title,
		Description: description,
		Price:       int64(price),
		PhotosUrls:  photoURL,
		Address:     address,
	}

	if err := h.Repo.Create(ad); err != nil {
		http.Error(w, "cannot save ad: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ad)
}

// Get возвращает объявление по его ID, но только если оно
// принадлежит пользователю с данным telegram_id.
// @Summary      Получить объявление
// @Description  Возвращает детали объявления по переданному идентификатору и telegram_id.
// @Tags         ads
// @Param        id            path      int  true  "ID объявления"
// @Param        telegram_id   query     int  true  "Telegram ID пользователя"
// @Success      200  {object}  domain.Advertisement
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /ads/{id} [get]
func (h *AdHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1) Парсим ID объявления из пути
	idStr := mux.Vars(r)["id"]
	adID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid ad id: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2) Берём telegram_id как строку
	telegramID := r.URL.Query().Get("telegram_id")
	if telegramID == "" {
		http.Error(w, "missing telegram_id", http.StatusBadRequest)
		return
	}

	// 3) Запрашиваем объявление с проверкой принадлежности
	ad, err := h.Repo.GetByIDAndTelegram(adID, telegramID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "ad not found or access denied", http.StatusNotFound)
		} else {
			log.Printf("GetByIDAndTelegram error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 4) Отдаём JSON
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

// ListByTelegram возвращает все активные объявления пользователя по его telegram_id.
// @Summary      Список объявлений пользователя
// @Description  Возвращает массив объявлений, принадлежащих пользователю с переданным telegram_id.
// @Tags         ads
// @Param        telegram_id   query     int  true  "Telegram ID пользователя"
// @Success      200  {array}   domain.Advertisement
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /ads [get]
func (h *AdHandler) ListByTelegram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1) Берём telegramId из path-параметра
	vars := mux.Vars(r)
	telegramID, ok := vars["telegramId"]
	if !ok || telegramID == "" {
		http.Error(w, "missing telegramId in path", http.StatusBadRequest)
		return
	}

	// 2) Запрашиваем объявления в репозитории
	ads, err := h.Repo.GetByTelegramID(telegramID)
	if err != nil {
		http.Error(w, "cannot fetch ads: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3) Если нет — возвращаем пустой массив
	if ads == nil {
		ads = []*domain.Advertisement{}
	}

	// 4) Сериализуем в JSON
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
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
