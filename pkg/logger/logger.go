package logger

import (
	"log"
	"os"
	"time"
)

// Level ログレベル
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// Logger ログ構造体
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	file        *os.File
}

// NewLogger 新しいロガーを作成
func NewLogger(logFilePath string) (*Logger, error) {
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		debugLogger: log.New(file, "[DEBUG] ", log.LstdFlags),
		infoLogger:  log.New(file, "[INFO] ", log.LstdFlags),
		warnLogger:  log.New(file, "[WARN] ", log.LstdFlags),
		errorLogger: log.New(file, "[ERROR] ", log.LstdFlags),
		file:        file,
	}, nil
}

// Debug デバッグログを出力
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
	// 標準出力にも出力（開発時）
	log.Printf("[DEBUG] "+format, v...)
}

// Info 情報ログを出力
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
	// 標準出力にも出力（開発時）
	log.Printf("[INFO] "+format, v...)
}

// Warn 警告ログを出力
func (l *Logger) Warn(format string, v ...interface{}) {
	l.warnLogger.Printf(format, v...)
	// 標準出力にも出力（開発時）
	log.Printf("[WARN] "+format, v...)
}

// Error エラーログを出力
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
	// 標準出力にも出力（開発時）
	log.Printf("[ERROR] "+format, v...)
}

// Close ロガーを閉じる
func (l *Logger) Close() error {
	return l.file.Close()
}

// WithContext コンテキスト付きログを出力
func (l *Logger) WithContext(context string) *ContextLogger {
	return &ContextLogger{
		logger:  l,
		context: context,
	}
}

// ContextLogger コンテキスト付きロガー
type ContextLogger struct {
	logger  *Logger
	context string
}

// Debug コンテキスト付きデバッグログ
func (cl *ContextLogger) Debug(format string, v ...interface{}) {
	cl.logger.Debug("[%s] "+format, append([]interface{}{cl.context}, v...)...)
}

// Info コンテキスト付き情報ログ
func (cl *ContextLogger) Info(format string, v ...interface{}) {
	cl.logger.Info("[%s] "+format, append([]interface{}{cl.context}, v...)...)
}

// Warn コンテキスト付き警告ログ
func (cl *ContextLogger) Warn(format string, v ...interface{}) {
	cl.logger.Warn("[%s] "+format, append([]interface{}{cl.context}, v...)...)
}

// Error コンテキスト付きエラーログ
func (cl *ContextLogger) Error(format string, v ...interface{}) {
	cl.logger.Error("[%s] "+format, append([]interface{}{cl.context}, v...)...)
}

// RequestLogger HTTPリクエストログ用
func (l *Logger) RequestLogger(method, path string, statusCode int, duration time.Duration, userID string) {
	l.Info("HTTP Request - Method: %s, Path: %s, Status: %d, Duration: %v, UserID: %s",
		method, path, statusCode, duration, userID)
} 