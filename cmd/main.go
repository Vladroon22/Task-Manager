package main

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "github.com/Vladroon22/TaskTracker/docs"
	"github.com/Vladroon22/TaskTracker/internal/database"
	"github.com/Vladroon22/TaskTracker/internal/handlers"
	repository "github.com/Vladroon22/TaskTracker/internal/repo"
	"github.com/Vladroon22/TaskTracker/internal/service"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Task Manager API
// @version         1.0
// @description     API для управления задачами и тегами
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@taskmanager.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @tag.name tasks
// @tag.description Управление задачами

// @tag.name tags
// @tag.description Управление тегами

// @tag.name tasks-period
// @tag.description Задачи с периодами выполнения

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	db, err := database.NewDB().Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tagRepo := repository.NewTagRepo(db)
	taskRepo := repository.NewTaskRepo(db, tagRepo)

	tagService := service.NewTagService(tagRepo)
	taskService := service.NewTaskService(taskRepo)

	tagHandler := handlers.NewTagHandler(tagService)
	taskHandler := handlers.NewTaskHandler(taskService)

	router := mux.NewRouter()

	// Swagger endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	// CRUD таски
	api.HandleFunc("/tasks", taskHandler.ListTasks).Methods("GET")
	api.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetTask).Methods("GET")
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.UpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.DeleteTask).Methods("DELETE")

	// CRUD теги
	api.HandleFunc("/tags", tagHandler.ListTags).Methods("GET")
	api.HandleFunc("/tags", tagHandler.CreateTag).Methods("POST")
	api.HandleFunc("/tags/{id:[0-9]+}", tagHandler.GetTag).Methods("GET")
	api.HandleFunc("/tags/{id:[0-9]+}", tagHandler.DeleteTag).Methods("DELETE")

	// Link between теги и задачи
	api.HandleFunc("/tasks/{taskId:[0-9]+}/tags", tagHandler.GetTaskTags).Methods("GET")
	api.HandleFunc("/tasks/{taskId:[0-9]+}/tags/{tagId:[0-9]+}", tagHandler.AddTagToTask).Methods("POST")
	api.HandleFunc("/tasks/{taskId:[0-9]+}/tags/{tagId:[0-9]+}", tagHandler.RemoveTagFromTask).Methods("DELETE")

	// Tasks with period
	api.HandleFunc("/tasks-period", taskHandler.CreateTaskWithPeriod).Methods("POST")
	api.HandleFunc("/tasks-period", taskHandler.GetTasksByPeriod).Methods("GET")
	api.HandleFunc("/tasks-period/{id:[0-9]+}", taskHandler.SetTaskStatus).Methods("PUT")
	api.HandleFunc("/tasks-period/{id:[0-9]+}", taskHandler.DeleteTaskOverride).Methods("DELETE")

	addr := os.Getenv("addr")
	log.Printf("Server starting on %s", addr)
	log.Printf("Swagger UI available at http://%s/swagger/index.html", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
