package handlers

import (
	"net/http"

	"github.com/Firebreather-heart/ningen/internal/config"
	"github.com/Firebreather-heart/ningen/internal/models"
	"github.com/gin-gonic/gin"
)

// Recommend handles POST /recommend.
// It accepts a user persona and returns contextual product recommendations.
func Recommend(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Persona models.Persona `json:"persona" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Invoke the recommender agent to generate recommendations
		c.JSON(http.StatusOK, gin.H{
			"message": "recommend endpoint",
			"status":  "not yet implemented",
		})
	}
}
