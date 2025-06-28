package handlers

import (
	"encoding/json"
	"net/http"
	"poppins/domain"
	"poppins/repository"
	"strconv"

	"github.com/gorilla/mux"
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
	id, _ := strconv.ParseInt(mux.Vars(r)["telegramId"], 10, 64)
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
	id, _ := strconv.ParseInt(mux.Vars(r)["telegramId"], 10, 64)
	u, err := h.Repo.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u)
}
