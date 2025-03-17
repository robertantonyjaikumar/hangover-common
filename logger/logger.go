package logger

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var zapLog *zap.Logger

func init() {
	var err error

	productionConfig := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	//zapcore.TimeEncoderOfLayout("Jan _2 15:04:05.000000000")
	//encoderConfig.StacktraceKey = "" // to hide stacktrace info
	productionConfig.EncoderConfig = encoderConfig

	zapLog, err = productionConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

}

type JwtSessionPayload struct {
	TID  string `json:"tid"  binding:"required"`
	Type string `json:"type" binding:"required"`
	RID  string `json:"rid"  binding:"required"`
	SID  string `json:"sid"  binding:"required"`
}

func GetZapLogger() *zap.Logger {
	return zapLog
}

func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}

func Panic(message string, fields ...zap.Field) {
	zapLog.Panic(message, fields...)
}
func DPanic(message string, fields ...zap.Field) {
	zapLog.Panic(message, fields...)
}
func InfoWithSessionCtx(ctx *gin.Context, message string, fields ...zap.Field) {
	generatedFields := generateFields(ctx, fields...)
	zapLog.Info(message, generatedFields...)
}

func DebugWithSessionCtx(ctx *gin.Context, message string, fields ...zap.Field) {
	generatedFields := generateFields(ctx, fields...)
	zapLog.Debug(message, generatedFields...)
}

func ErrorWithSessionCtx(ctx *gin.Context, message string, fields ...zap.Field) {
	generatedFields := generateFields(ctx, fields...)
	zapLog.Error(message, generatedFields...)
}

func FatalWithSessionCtx(ctx *gin.Context, message string, fields ...zap.Field) {
	generatedFields := generateFields(ctx, fields...)
	zapLog.Fatal(message, generatedFields...)
}

func generateFields(ctx *gin.Context, fields ...zap.Field) []zap.Field {
	if ctx != nil {
		claimPayload, exists := ctx.Get("x-claim-payload-log")
		if !exists {
			return fields
		}

		jwtSessionPayload, _ := claimPayload.(JwtSessionPayload)

		contextField := zap.Any("session_id", jwtSessionPayload.SID)
		return append([]zap.Field{contextField}, fields...)
	}
	return nil
}
