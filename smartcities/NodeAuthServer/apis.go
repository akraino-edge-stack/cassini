package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func middlewareLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

func middlewareRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// curl -v -X GET 127.0.0.1:8301/nodes
func GetNodes(c *gin.Context) {
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	} else {
		c.String(http.StatusOK, string(output))
	}
}

// curl -v -X GET 127.0.0.1:8301/pods
func GetPods(c *gin.Context) {
	cmd := exec.Command("kubectl", "get", "pods", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	} else {
		c.String(http.StatusOK, string(output))
	}
}

func StartGinApis() {
	gin.SetMode(gin.ReleaseMode) // set to release mode

	r := gin.New()
	r.Use(middlewareLogger(zap.L()))
	r.Use(middlewareRecovery(zap.L(), true))

	// use for check deamon was runned
	// curl 127.0.0.1:8301/version
	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, "1.0")
	})

	r.GET("/nodes", GetNodes)
	r.GET("/pods", GetPods)

	r.Run(fmt.Sprintf(":%d", 8301))
}
