package handlers

import (
	"net/http"

	"github.com/Firebreather-heart/ningen/internal/config"
	"github.com/Firebreather-heart/ningen/internal/models"
	"github.com/gin-gonic/gin"
)

// SimulateReview handles POST /simulate-review.
// It accepts a user persona and product details, then generates a simulated review.
func SimulateReview(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Persona models.Persona `json:"persona" binding:"required"`
			Product models.Product `json:"product" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Invoke the modeling agent to generate a simulated review
		c.JSON(http.StatusOK, gin.H{
			"message": "simulate-review endpoint",
			"status":  "not yet implemented",
		})
	}
}
