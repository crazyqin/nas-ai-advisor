package aiadvisor

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// API represents the REST API server
type API struct {
	advisor *Advisor
	router  *gin.Engine
}

// NewAPI creates a new API server
func NewAPI(advisor *Advisor) *API {
	api := &API{
		advisor: advisor,
		router:  gin.Default(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes sets up the API routes
func (a *API) setupRoutes() {
	// Health check
	a.router.GET("/health", a.healthCheck)

	// API v1
	v1 := a.router.Group("/api/v1")
	{
		// Query endpoint
		v1.POST("/query", a.query)

		// Recommendations endpoint
		v1.POST("/recommendations", a.getRecommendations)

		// Diagnostic endpoint
		v1.POST("/diagnose", a.diagnose)

		// System status endpoint
		v1.GET("/status", a.getStatus)

		// Knowledge base endpoints
		knowledge := v1.Group("/knowledge")
		{
			knowledge.GET("/articles", a.getArticles)
			knowledge.GET("/articles/:id", a.getArticle)
			knowledge.POST("/articles", a.createArticle)
			knowledge.DELETE("/articles/:id", a.deleteArticle)
			knowledge.GET("/categories", a.getCategories)
		}

		// Conversation endpoints
		conversations := v1.Group("/conversations")
		{
			conversations.GET("/:id", a.getConversation)
			conversations.GET("/:id/messages", a.getConversationMessages)
			conversations.DELETE("/:id", a.deleteConversation)
		}
	}

	// Serve static files
	a.router.Static("/static", "./web/static")
	a.router.StaticFile("/", "./web/static/index.html")
}

// healthCheck handles health check requests
func (a *API) healthCheck(c *gin.Context) {
	health := a.advisor.HealthCheck()
	c.JSON(http.StatusOK, health)
}

// query handles query requests
func (a *API) query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	resp, err := a.advisor.Query(ctx, &req)
	if err != nil {
		log.Printf("Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getRecommendations handles recommendation requests
func (a *API) getRecommendations(c *gin.Context) {
	var req RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	recommendations, err := a.advisor.GetRecommendations(ctx, &req)
	if err != nil {
		log.Printf("Recommendations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

// diagnose handles diagnostic requests
func (a *API) diagnose(c *gin.Context) {
	var req DiagnosticRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	result, err := a.advisor.Diagnose(ctx, &req)
	if err != nil {
		log.Printf("Diagnostic error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform diagnosis"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getStatus handles status requests
func (a *API) getStatus(c *gin.Context) {
	status := a.advisor.GetStatus()
	c.JSON(http.StatusOK, status)
}

// getArticles handles get articles requests
func (a *API) getArticles(c *gin.Context) {
	articles := a.advisor.knowledge.GetAll()
	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

// getArticle handles get article requests
func (a *API) getArticle(c *gin.Context) {
	id := c.Param("id")
	article, ok := a.advisor.knowledge.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

// createArticle handles create article requests
func (a *API) createArticle(c *gin.Context) {
	var article KnowledgeArticle
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.advisor.knowledge.Add(&article); err != nil {
		log.Printf("Create article error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, article)
}

// deleteArticle handles delete article requests
func (a *API) deleteArticle(c *gin.Context) {
	id := c.Param("id")
	if err := a.advisor.knowledge.Delete(id); err != nil {
		log.Printf("Delete article error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}

// getCategories handles get categories requests
func (a *API) getCategories(c *gin.Context) {
	categories := a.advisor.knowledge.GetCategories()
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// getConversation handles get conversation requests
func (a *API) getConversation(c *gin.Context) {
	id := c.Param("id")
	conversation, ok := a.advisor.conversations.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// getConversationMessages handles get conversation messages requests
func (a *API) getConversationMessages(c *gin.Context) {
	id := c.Param("id")
	messages, err := a.advisor.conversations.GetMessages(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// deleteConversation handles delete conversation requests
func (a *API) deleteConversation(c *gin.Context) {
	id := c.Param("id")
	if err := a.advisor.conversations.Delete(id); err != nil {
		log.Printf("Delete conversation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

// Run starts the API server
func (a *API) Run(addr string) error {
	return a.router.Run(addr)
}

// RunTLS starts the API server with TLS
func (a *API) RunTLS(addr, certFile, keyFile string) error {
	return a.router.RunTLS(addr, certFile, keyFile)
}

// GetRouter returns the gin router for testing
func (a *API) GetRouter() *gin.Engine {
	return a.router
}

// JSONMiddleware is a middleware for JSON content type
func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

// CORSMiddleware is a middleware for CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware is a middleware for request logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[API] %s %s %d %v", method, path, status, latency)
	}
}

// ResponseWriter wraps gin.ResponseWriter for response capture
type ResponseWriter struct {
	gin.ResponseWriter
	body []byte
}

// Write captures the response body
func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.body = b
	return w.ResponseWriter.Write(b)
}

// WriteString captures the response string
func (w *ResponseWriter) WriteString(s string) (int, error) {
	w.body = []byte(s)
	return w.ResponseWriter.WriteString(s)
}

// GetBody returns the captured response body
func (w *ResponseWriter) GetBody() []byte {
	return w.body
}

// MarshalJSON marshals the response to JSON
func (w *ResponseWriter) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.body)
}
