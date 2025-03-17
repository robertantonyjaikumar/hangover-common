package middlewares

import (
	"bytes"
	"encoding/json"
	ginzap "github.com/gin-contrib/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func LogResponseAndRequestBodyMiddleware(logger *zap.Logger, conf *ginzap.Config) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		if _, ok := skipPaths[path]; !ok {
			var fields []zapcore.Field
			request, _ := ioutil.ReadAll(c.Request.Body)
			reqBodyCloser := ioutil.NopCloser(bytes.NewBuffer(request))
			var requestBody interface{}
			json.Unmarshal([]byte(request), &requestBody)
			fields = []zapcore.Field{
				zap.Any("request-body", requestBody),
			}
			c.Request.Body = reqBodyCloser
			blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
			c.Next()
			var responseBody interface{}
			response, _ := ioutil.ReadAll(blw.body)
			json.Unmarshal([]byte(response), &responseBody)
			fields = append(fields, zap.Any("response-body", responseBody))
			end := time.Now()
			latency := end.Sub(start)
			if conf.UTC {
				end = end.UTC()
			}
			fields = append(fields, zap.Int("status", c.Writer.Status()))
			fields = append(fields, zap.String("method", c.Request.Method))
			fields = append(fields, zap.String("path", path))
			fields = append(fields, zap.String("query", query))
			fields = append(fields, zap.String("ip", c.ClientIP()))
			fields = append(fields, zap.String("user-agent", c.Request.UserAgent()))
			fields = append(fields, zap.Duration("latency", latency))

			if conf.TimeFormat != "" {
				fields = append(fields, zap.String("time", end.Format(conf.TimeFormat)))
			}
			//if conf.TraceID {
			//	fields = append(
			//		fields,
			//		zap.String(
			//			"traceID",
			//			trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String(),
			//		),
			//	)
			//}

			if conf.Context != nil {
				fields = append(fields, conf.Context(c)...)
			}

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				for _, e := range c.Errors.Errors() {
					logger.Error(e, fields...)
				}
			} else {
				logger.Info(path, fields...)
			}
		}
	}
}
