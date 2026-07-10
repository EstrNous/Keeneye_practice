package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func InstanceID() gin.HandlerFunc {
	id := resolveInstanceID()
	return func(c *gin.Context) {
		c.Header("X-Backend-Instance", id)
		c.Set("instanceID", id)
		c.Next()
	}
}

func resolveInstanceID() string {
	if id := os.Getenv("INSTANCE_ID"); id != "" {
		return id
	}
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "api/backend_unknown"
	}
	idx := strings.LastIndex(hostname, "-")
	if idx < 0 || idx == len(hostname)-1 {
		return "api/" + hostname
	}
	return "api/backend_" + hostname[idx+1:]
}
