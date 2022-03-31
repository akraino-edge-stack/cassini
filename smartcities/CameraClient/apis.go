package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
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

// curl 127.0.0.1:8303/camera
func GetCameraImage(c *gin.Context) {
	// create tmp dir
	tmpPath := "/tmp/CameraClient" + RandomString(8)
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		os.Mkdir(tmpPath, 0777)
	}
	// run cmd in tmp dir
	cmd := exec.Command("nvgstcapture", "-A", "-C", "1", "-S", "1", "--capture-auto", "--image-res=2")
	// cmd := exec.Command("bash", "-c", "echo \"Hello\" > test.txt")
	cmd.Dir = tmpPath
	_, err := cmd.CombinedOutput()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	} else {
		// get captured img in tmp dir
		files, err := ioutil.ReadDir(tmpPath)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		} else {
			for _, f := range files {
				if !f.IsDir() {
					inPath := filepath.Join(tmpPath, f.Name())
					buffer, _ := ioutil.ReadFile(inPath)
					c.Data(http.StatusOK, "application/octet-stream", buffer)
					break
				}
			}
		}
	}
	// remove tmp dir
	os.RemoveAll(tmpPath)
}

func StartGinApis() {
	gin.SetMode(gin.ReleaseMode) // set to release mode

	r := gin.New()
	r.Use(middlewareLogger(zap.L()))
	r.Use(middlewareRecovery(zap.L(), true))

	// use for check deamon was runned
	// curl 128.0.0.1:8303/version
	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, "1.0")
	})

	r.GET("/camera", GetCameraImage)

	r.Run(fmt.Sprintf(":%d", 8303))
}
