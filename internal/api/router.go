package api

import (
	"github.com/Firebreather-heart/ningen/internal/api/handlers"
	"github.com/Firebreather-heart/ningen/internal/config"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin engine with all route definitions.
func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/simulate-review", handlers.SimulateReview(cfg))
		v1.POST("/recommend", handlers.Recommend(cfg))
	}

	return r
}
