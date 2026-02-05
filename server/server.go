package server

import (
	"asynqdemo/tasks"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// Server represents the HTTP server
type Server struct {
	client *asynq.Client
	router *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(redisAddr string) *Server {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	router := gin.Default()

	s := &Server{
		client: client,
		router: router,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		api.GET("/health", s.healthCheck)

		taskGroup := api.Group("/tasks")
		{
			taskGroup.POST("/email", s.submitEmailTask)
			taskGroup.POST("/process", s.submitDataProcessTask)
		}
	}
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// EmailTaskRequest represents the request body for email tasks
type EmailTaskRequest struct {
	To      string `json:"to" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

// submitEmailTask handles email task submission
func (s *Server) submitEmailTask(c *gin.Context) {
	var req EmailTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := tasks.NewEmailDeliveryTask(req.To, req.Subject, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	info, err := s.client.Enqueue(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}

	log.Printf("📨 Enqueued email task: ID=%s Queue=%s", info.ID, info.Queue)

	c.JSON(http.StatusOK, gin.H{
		"task_id": info.ID,
		"queue":   info.Queue,
		"message": "Email task submitted successfully",
	})
}

// DataProcessTaskRequest represents the request body for data processing tasks
type DataProcessTaskRequest struct {
	DataID string `json:"data_id" binding:"required"`
	Action string `json:"action" binding:"required"`
	Delay  int    `json:"delay"`
}

// submitDataProcessTask handles data processing task submission
func (s *Server) submitDataProcessTask(c *gin.Context) {
	var req DataProcessTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default delay to 2 seconds if not specified
	if req.Delay == 0 {
		req.Delay = 2
	}

	task, err := tasks.NewDataProcessTask(req.DataID, req.Action, req.Delay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Enqueue with options (e.g., process in 5 seconds)
	info, err := s.client.Enqueue(task, asynq.ProcessIn(5*time.Second))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}

	log.Printf("📨 Enqueued data process task: ID=%s Queue=%s", info.ID, info.Queue)

	c.JSON(http.StatusOK, gin.H{
		"task_id":    info.ID,
		"queue":      info.Queue,
		"message":    "Data processing task submitted successfully",
		"process_at": "in 5 seconds",
	})
}

// Run starts the HTTP server
func (s *Server) Run(addr string) error {
	log.Printf("🌐 Starting HTTP server on %s", addr)
	return s.router.Run(addr)
}

// Close closes the Asynq client
func (s *Server) Close() error {
	return s.client.Close()
}
