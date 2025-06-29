package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"poppins/domain"
	"poppins/repository"
)

type UserHandler struct {
	Repo *repository.UserRepo
}

func NewUserHandler(repo *repository.UserRepo) *UserHandler {
	return &UserHandler{Repo: repo}
}

// Create создаёт нового пользователя.
// @Summary      Создать пользователя
// @Description  Принимает JSON с данными пользователя и сохраняет его в БД.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      domain.User  true  "Данные пользователя"
// @Success      201   {object}  domain.User
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var u domain.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Repo.Create(&u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

// Get возвращает пользователя по TelegramID.
// @Summary      Получить пользователя
// @Description  Возвращает пользователя из БД по переданному в пути идентификатору.
// @Tags         users
// @Produce      json
// @Param        telegramId   path      int  true  "ID пользователя"
// @Success      200  {object}  domain.User
// @Failure      404  {object}  map[string]string
// @Router       /users/{telegramId} [get]
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := mux.Vars(r)["telegramId"]
	u, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u)
}

// Delete удаляет пользователя по TelegramID.
// @Summary      Удалить пользователя
// @Description  Удаляет запись пользователя из БД по его идентификатору.
// @Tags         users
// @Produce      json
// @Param        telegramId   path      int  true  "TelegramID пользователя"
// @Success      200  {object}  domain.User
// @Failure      404  {object}  map[string]string
// @Router       /users/{telegramId} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := mux.Vars(r)["telegramId"]
	u, err := h.Repo.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u)
}

type UpdateNameRequest struct {
	Name string `json:"name"`
}

// PATCH /users/{telegramId}/name
func (h *UserHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Парсим telegramId
	id := mux.Vars(r)["telegramId"]

	// 2. Декодим тело { "name": "Новое Имя" }
	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Обновляем в репозитории
	u, err := h.Repo.UpdateName(id, req.Name)
	if err != nil {
		log.Printf("UpdateName error: %v", err)
		http.Error(w, "could not update name", http.StatusInternalServerError)
		return
	}

	// 4. Возвращаем обновлённого пользователя
	json.NewEncoder(w).Encode(u)
}

// UpdatePhoneRequest — payload для изменения телефона
type UpdatePhoneRequest struct {
	Phone string `json:"phone"`
}

// PATCH /users/{telegramId}/phone
func (h *UserHandler) UpdatePhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["telegramId"]

	var req UpdatePhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	u, err := h.Repo.UpdatePhone(id, req.Phone)
	if err != nil {
		log.Printf("UpdatePhone error: %v", err)
		http.Error(w, "could not update phone", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(u)
}

// UpdateContactRequest — payload для изменения способа связи
type UpdateContactRequest struct {
	PreferredContact string `json:"preferred_contact"`
}

func (h *UserHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1) Достаём telegramId из пути
	vars := mux.Vars(r)
	telegramID, ok := vars["telegramId"]
	if !ok || telegramID == "" {
		http.Error(w, "missing telegramId in path", http.StatusBadRequest)
		return
	}

	// 2) Декодим тело запроса
	var req UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3) Вызываем репозиторий
	u, err := h.Repo.UpdatePreferredContact(telegramID, string(req.PreferredContact))
	if err != nil {
		log.Printf("UpdatePreferredContact error: %v", err)
		http.Error(w, "could not update preferred_contact", http.StatusInternalServerError)
		return
	}

	// 4) Отдаём обновлённого пользователя
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}
