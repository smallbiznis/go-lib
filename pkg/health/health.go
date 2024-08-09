package health

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	mu sync.Mutex
)

// HealthStatus represents the application's health status
type HealthStatus struct {
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

func LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		healthStatus := HealthStatus{Status: "alive"}
		c.JSON(http.StatusOK, healthStatus)
	}
}

func ReadinessHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		if err := db.Ping(); err == nil {
			healthStatus := HealthStatus{Status: "ready"}
			c.JSON(http.StatusOK, healthStatus)
		} else {
			healthStatus := HealthStatus{Status: "unavailable", Details: "database connection issue"}
			c.JSON(http.StatusServiceUnavailable, healthStatus)
		}
	}
}
