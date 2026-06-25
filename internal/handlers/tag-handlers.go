package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Vladroon22/TaskTracker/internal/models"
	"github.com/Vladroon22/TaskTracker/internal/service"
	"github.com/gorilla/mux"
)

type TagHandler struct {
	srv service.TagServicer
}

func NewTagHandler(srv service.TagServicer) *TagHandler {
	return &TagHandler{srv: srv}
}

// @Summary      Создать тег
// @Description  Создание нового тега
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateTagRequest  true  "Данные тега"
// @Success      201      {object}  models.Tag
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tags [post]
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Name) == 0 {
		respondWithError(w, http.StatusBadRequest, "Tag name is required")
		return
	}

	tag, err := h.srv.CreateTag(r.Context(), req.Name)
	if err != nil {
		log.Printf("Error creating tag: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, tag)
}

// @Summary      Список тегов
// @Description  Получение списка всех тегов
// @Tags         tags
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Tag
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tags [get]
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.srv.ListTags(r.Context())
	if err != nil {
		log.Printf("Error listing tags: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to list tags")
		return
	}

	if tags == nil {
		tags = []models.Tag{}
	}

	respondWithJSON(w, http.StatusOK, tags)
}

// @Summary      Получить тег по ID
// @Description  Получение тега по идентификатору
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID тега"
// @Success      200  {object}  models.Tag
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tags/{id} [get]
func (h *TagHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	tag, err := h.srv.GetTagByID(r.Context(), id)
	if err != nil {
		if err == models.ErrTagNotFound {
			respondWithError(w, http.StatusNotFound, "Tag not found")
			return
		}
		log.Printf("Error getting tag: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get tag")
		return
	}

	respondWithJSON(w, http.StatusOK, tag)
}

// @Summary      Удалить тег
// @Description  Удаление тега по ID
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID тега"
// @Success      200  {object}  models.MessageResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tags/{id} [delete]
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	if err := h.srv.DeleteTag(r.Context(), id); err != nil {
		if err == models.ErrTagIsSystem {
			respondWithError(w, http.StatusForbidden, "Cannot delete system tag")
			return
		}
		if err == models.ErrTagNotFound {
			respondWithError(w, http.StatusNotFound, "Tag not found")
			return
		}
		log.Printf("Error deleting tag: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Tag deleted successfully"})
}

// @Summary      Добавить тег к задаче
// @Description  Привязка тега к задаче
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        taskId   path      int  true  "ID задачи"
// @Param        tagId    path      int  true  "ID тега"
// @Param        request  body      models.CreateTagRequest  false  "Данные тега (опционально)"
// @Success      200      {object}  models.MessageResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks/{taskId}/tags/{tagId} [post]
func (h *TagHandler) AddTagToTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["taskId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	tagID, err := strconv.Atoi(vars["tagId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	if err := h.srv.AddTagToTask(r.Context(), taskID, tagID, req.Name); err != nil {
		if err == models.ErrTaskNotFound || err == models.ErrTagNotFound {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("Error adding tag to task: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to add tag to task")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Tag added to task successfully"})
}

// @Summary      Удалить тег из задачи
// @Description  Отвязка тега от задачи
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        taskId  path      int  true  "ID задачи"
// @Param        tagId   path      int  true  "ID тега"
// @Success      200     {object}  models.MessageResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks/{taskId}/tags/{tagId} [delete]
func (h *TagHandler) RemoveTagFromTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["taskId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	tagID, err := strconv.Atoi(vars["tagId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	if err := h.srv.RemoveTagFromTask(r.Context(), taskID, tagID); err != nil {
		if err == models.ErrTagNotOnTask {
			respondWithError(w, http.StatusNotFound, "Tag not found on task")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to remove tag from task")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Tag removed from task successfully"})
}

// @Summary      Получить теги задачи
// @Description  Получение списка тегов, привязанных к задаче
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        taskId  path      int  true  "ID задачи"
// @Success      200     {array}   models.Tag
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks/{taskId}/tags [get]
func (h *TagHandler) GetTaskTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["taskId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	tags, err := h.srv.GetTaskTags(r.Context(), taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Tags not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get task tags")
		return
	}

	if tags == nil {
		tags = []models.Tag{}
	}

	respondWithJSON(w, http.StatusOK, tags)
}
