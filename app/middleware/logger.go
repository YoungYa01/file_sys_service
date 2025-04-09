package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// 请求前
		c.Next()
		// 请求后
		latency := time.Since(t)
		fmt.Print("Time Consuming is: ", latency)
	}
}
