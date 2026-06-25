package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Vladroon22/TaskTracker/internal/models"
	"github.com/Vladroon22/TaskTracker/internal/service"
	"github.com/gorilla/mux"
)

type TaskHandler struct {
	srv service.TaskServicer
}

func NewTaskHandler(srv service.TaskServicer) *TaskHandler {
	return &TaskHandler{srv: srv}
}

// @Summary      Создать задачу
// @Description  Создание новой задачи
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateTaskRequest  true  "Данные задачи"
// @Success      201      {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { /// <-----
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.DueDate.IsZero() {
		respondWithError(w, http.StatusBadRequest, "Due date is required")
		return
	}

	task, err := h.srv.Create(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

// @Summary      Получить задачу по ID
// @Description  Получение задачи по идентификатору
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID задачи"
// @Success      200  {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.srv.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Println(err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

// @Summary      Список задач с фильтрацией
// @Description  Получение списка задач с возможностью фильтрации и пагинации
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        status     query     string  false  "Статус задачи (new, in_progress, done, cancelled)"
// @Param        date_from  query     string  false  "Дата начала периода (YYYY-MM-DD)"
// @Param        date_to    query     string  false  "Дата конца периода (YYYY-MM-DD)"
// @Param        tags       query     []string  false  "Теги для фильтрации"
// @Param        limit      query     int     false  "Лимит записей"        default(10)
// @Param        offset     query     int     false  "Смещение для пагинации" default(0)
// @Success      200        {object}  models.TaskListResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks [get]
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	filter := &models.TaskFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		s := models.TaskStatus(status)
		if s.IsValid() {
			filter.Status = &s
		}
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.DateOnly, dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse(time.DateOnly, dateTo); err == nil {
			filter.DateTo = &t
		}
	}

	if tags := r.URL.Query()["tags"]; len(tags) > 0 {
		filter.Tags = tags
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	tasks, err := h.srv.List(r.Context(), filter)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list tasks")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

// @Summary      Обновить задачу
// @Description  Обновление существующей задачи
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id       path      int                        true  "ID задачи"
// @Param        request  body      models.UpdateTaskRequest   true  "Обновленные данные задачи"
// @Success      200      {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // <---
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := h.srv.Update(r.Context(), id, &req)
	if err != nil {
		if err == models.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

// @Summary      Удалить задачу
// @Description  Удаление задачи по ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID задачи"
// @Success      200  {object}  models.MessageResponse
// / @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if err := h.srv.Delete(r.Context(), id); err != nil {
		if err == models.ErrTaskNotFound {
			respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

// @Summary      Создать задачу с периодом
// @Description  Создание задачи с указанием периода выполнения
// @Tags         tasks-period
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateTaskPeriodRequest  true  "Данные задачи с периодом"
// @Success      201      {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks-period [post]
func (h *TaskHandler) CreateTaskWithPeriod(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskPeriodRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.DueDate.Time().IsZero() {
		respondWithError(w, http.StatusBadRequest, "Due date is required")
		return
	}

	task, err := h.srv.CreateTaskWithPeriod(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

// @Summary      Получить задачи за период
// @Description  Получение списка задач за указанный период
// @Tags         tasks-period
// @Accept       json
// @Produce      json
// @Param        from  query     string  true  "Начало периода (YYYY-MM-DD)"
// @Param        to    query     string  true  "Конец периода (YYYY-MM-DD)"
// @Success      200   {object}  models.TaskListResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks-period [get]
func (h *TaskHandler) GetTasksByPeriod(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		respondWithError(w, http.StatusBadRequest, "from and to parameters are required")
		return
	}

	from, err := time.Parse(time.DateOnly, fromStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid 'from' date format. Use YYYY-MM-DD")
		return
	}

	to, err := time.Parse(time.DateOnly, toStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid 'to' date format. Use YYYY-MM-DD")
		return
	}

	to = to.Add(24*time.Hour - 1*time.Second)

	tasks, err := h.srv.GetTasksForPeriod(r.Context(), from, to)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

// @Summary      Установить статус задачи на дату
// @Description  Установка статуса задачи на конкретную дату
// @Tags         tasks-period
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "ID задачи"
// @Param        date    query     string  true  "Дата (YYYY-MM-DD)"
// @Param        request body      models.SetStatusRequest  true  "Статус задачи"
// @Success      200     {object}  models.MessageResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks-period/{id} [put]
func (h *TaskHandler) SetTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	date, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	status := models.TaskStatus(req.Status)

	validStatuses := map[models.TaskStatus]bool{
		models.StatusNew:        true,
		models.StatusInProgress: true,
		models.StatusDone:       true,
		models.StatusCancelled:  true,
	}

	if !validStatuses[status] {
		respondWithError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.srv.SetTaskStatus(r.Context(), id, date, status); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Status updated successfully"})
}

// @Summary      Удалить переопределение статуса задачи
// @Description  Удаление переопределения статуса задачи на конкретную дату
// @Tags         tasks-period
// @Accept       json
// @Produce      json
// @Param        id    path      int     true  "ID задачи"
// @Param        date  query     string  true  "Дата переопределения (YYYY-MM-DD)"
// @Success      200   {object}  models.MessageResponse
// @Failure      400   {object}  models.ErrorResponse  "Некорректный ID задачи или формат даты"
// @Failure      404   {object}  models.BadRequestResponse  "Переопределение не найдено"
// @Failure      500   {object}  models.ServerErrorResponse  "Внутренняя ошибка сервера"
// @Router       /tasks-period/{id} [delete]
func (h *TaskHandler) DeleteTaskOverride(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	date, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	if err := h.srv.DeleteTaskOverride(r.Context(), id, date); err != nil {
		if strings.Contains(err.Error(), "override not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Override for task %d on %s deleted successfully", id, dateStr),
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"Error": message})
}
