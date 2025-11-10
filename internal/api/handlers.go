package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourusername/go-rag/internal/models"
	"github.com/yourusername/go-rag/internal/service"
)

// DocumentRequest represents a request to store a document
type DocumentRequest struct {
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchRequest represents a request to search for similar documents
type SearchRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit,omitempty"`
}

// Server represents the HTTP server for the RAG API
type Server struct {
	router     *gin.Engine
	ragService service.RAGService
}

// NewServer creates a new API server
func NewServer(ragService service.RAGService) *Server {
	if ragService == nil {
		panic("RAG service is required")
	}

	router := gin.Default()
	server := &Server{
		router:     router,
		ragService: ragService,
	}

	server.SetupRoutes(router)
	return server
}

// SetupRoutes configures the HTTP routes for the server
func (s *Server) SetupRoutes(router *gin.Engine) {
	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Document routes
		documents := api.Group("/documents")
		{
			documents.POST("", s.StoreDocumentHandler)
			documents.GET("/:id", s.GetDocumentHandler)
			documents.GET("", s.ListDocumentsHandler)
			documents.DELETE("/:id", s.DeleteDocumentHandler)
		}

		// Search route
		api.POST("/search", s.SearchHandler)

		// RAG query route
		api.POST("/query", s.QueryHandler)
	}
}

// Start starts the HTTP server
func (s *Server) Start(port string) error {
	return s.router.Run(":" + port)
}

// StoreDocumentHandler handles document storage requests
func (s *Server) StoreDocumentHandler(c *gin.Context) {
	var request DocumentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	documentID, err := s.ragService.AddDocument(c.Request.Context(), request.Content, request.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store document: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": documentID})
}

// GetDocumentHandler handles document retrieval requests
func (s *Server) GetDocumentHandler(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	_, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID format"})
		return
	}

	// This requires a direct reference to the database, which the current design doesn't expose.
	// In a real application, you would add a GetDocument method to the RAGService interface.
	c.JSON(http.StatusOK, gin.H{"message": "Document retrieval not implemented in MVP", "id": idParam})
}

// ListDocumentsHandler handles requests to list documents
func (s *Server) ListDocumentsHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// This requires a direct reference to the database, which the current design doesn't expose.
	// In a real application, you would add a ListDocuments method to the RAGService interface.
	c.JSON(http.StatusOK, gin.H{"message": "Document listing not implemented in MVP"})
}

// DeleteDocumentHandler handles document deletion requests
func (s *Server) DeleteDocumentHandler(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	_, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID format"})
		return
	}

	// This requires a direct reference to the database, which the current design doesn't expose.
	// In a real application, you would add a DeleteDocument method to the RAGService interface.
	c.JSON(http.StatusOK, gin.H{"message": "Document deletion not implemented in MVP", "id": idParam})
}

// SearchHandler handles vector similarity search requests
func (s *Server) SearchHandler(c *gin.Context) {
	var request SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	results, err := s.ragService.SearchSimilar(c.Request.Context(), request.Query, request.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// QueryHandler handles RAG query requests
func (s *Server) QueryHandler(c *gin.Context) {
	var request models.RAGQuery
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := s.ragService.Query(c.Request.Context(), request.Query, request.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
