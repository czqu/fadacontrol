package logger

import (
	"errors"
	"fadacontrol/internal/base/conf"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"sync"
	"time"
)
import "gopkg.in/natefinch/lumberjack.v2"

var logger *Logger

type Loglevel uint32

const (
	Unknown Loglevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	END
)

func (l Loglevel) isValid() bool {
	return l > Unknown && l < END

}
func (l Loglevel) zapLevel() zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

type CircularBuffer struct {
	buffer        chan string
	size          int
	output        map[int]chan string
	mu            sync.Mutex
	broadcastOnce sync.Once
	readTimeout   time.Duration
}

func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		buffer: make(chan string, size),
		size:   size,
		output: make(map[int]chan string, 5),
	}

}

func (cb *CircularBuffer) Write(p []byte) (n int, err error) {
	value := string(p)

	select {
	case cb.buffer <- value:
	default:
		<-cb.buffer
		cb.buffer <- value
	}

	return len(p), nil
}

func (cb *CircularBuffer) broadcast() {
	for val := range cb.buffer {
		cb.mu.Lock()

		for _, ch := range cb.output {
			select {

			case ch <- val:

			case <-time.After(cb.readTimeout):

			}

		}

		cb.mu.Unlock()
	}
}
func (cb *CircularBuffer) RemoveReader(id int) {

	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.output, id)
}
func (cb *CircularBuffer) AddReader(reader chan string) int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	id := len(cb.output)
	cb.output[id] = reader
	cb.broadcastOnce.Do(func() {
		go cb.broadcast()
	})

	//fmt.Println("add reader: ", id)
	return id

}

type Logger struct {
	logger       *zap.Logger
	sugar        *zap.SugaredLogger
	level        Loglevel
	r            *conf.Conf
	logPath      string
	logLevel     string
	cyBuffer     *CircularBuffer
	cyBufferSize int
}

var once sync.Once

func GetBuffer() *CircularBuffer {
	if logger == nil {
		return nil
	}
	return logger.cyBuffer
}
func NewLogger(r *conf.Conf) *Logger {
	return &Logger{r: r}
}
func (l *Logger) GetBufferSize() int {
	return l.cyBufferSize
}
func InitLog(c *conf.Conf) {
	once.Do(func() {
		if logger != nil {
			return
		}
		l := new(Logger)
		logger = l
		l.r = c
		l.cyBufferSize = 100
		l.cyBuffer = NewCircularBuffer(l.cyBufferSize)
		l.cyBuffer.readTimeout = time.Second * 10
		var err error
		logger.logLevel = l.r.LogLevel
		logger.logPath, err = filepath.Abs(l.r.GetWorkdir() + "/log/" + l.r.LogName)
		if err != nil {
			logger = nil
			return
		}
		err = l.Init(l.logPath, str2Loglevel(l.logLevel))
		if err != nil {
			logger = nil
			return
		}
		return
	})

}
func GetLogPath() string {
	if logger == nil {
		return ""
	}
	return logger.logPath
}
func GetLogLevel() string {
	if logger == nil {
		return ""
	}
	return logger.logLevel
}

func (l *Logger) Init(logPath string, loglevel Loglevel) error {

	_, err := os.Stat(logPath)

	if err != nil {
		if os.IsNotExist(err) {

			err = os.MkdirAll(filepath.Dir(logPath), os.ModePerm)
			if err != nil {

				return err
			}
		} else if os.IsPermission(err) {

			return err
		} else {

			return err
		}
	}

	if !loglevel.isValid() {

		return errors.New("invalid log level")
	}
	l.level = loglevel
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(l.level.zapLevel())
	logHook := lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    1,
		MaxBackups: 100,
		MaxAge:     30,
		Compress:   false,
	}
	flushInterval := 30 * time.Second
	if logger.r.Debug == true {
		flushInterval = 1 * time.Second
	}
	bufferedWriteSyncer := zapcore.AddSync(&zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(&logHook),
		Size:          256 * 1024, // 256 KB buffer size
		FlushInterval: flushInterval,
	})
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(bufferedWriteSyncer, zapcore.AddSync(os.Stdout), zapcore.AddSync(l.cyBuffer)),
		atomicLevel,
	)

	var coreArr []zapcore.Core

	coreArr = append(coreArr, core)
	if loglevel == DebugLevel {
		l.logger = zap.New(zapcore.NewTee(coreArr...), zap.AddCallerSkip(2), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		l.logger = zap.New(zapcore.NewTee(coreArr...), zap.AddCallerSkip(2), zap.AddCaller())
	}

	l.sugar = l.logger.Sugar()
	return nil
}

func (l *Logger) Sync() {

	l.logger.Sync()
}
func (l *Logger) Warn(args ...interface{}) {

	l.sugar.Warn(args)
}
func (l *Logger) Debug(args ...interface{}) {

	l.sugar.Debug(args)
}
func (l *Logger) Info(args ...interface{}) {

	l.sugar.Info(args)
}
func (l *Logger) Error(args ...interface{}) {

	l.sugar.Error(args)
}

var loglevel Loglevel = InfoLevel

func str2Loglevel(level string) Loglevel {
	switch level {
	case "debug":
		loglevel = DebugLevel
		break
	case "info":
		loglevel = InfoLevel
		break
	case "warn":
		loglevel = WarnLevel
		break
	case "error":
		loglevel = ErrorLevel
		break
	default:
		loglevel = InfoLevel

	}
	return loglevel
}

func (l *Logger) Debugf(format string, v ...interface{}) {

	l.sugar.Debugf(format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {

	l.sugar.Infof(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {

	l.sugar.Errorf(format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {

	l.sugar.Warnf(format, v...)
}

func Sync() {
	logger.Sync()

}

func Warn(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		return
	}
	logger.Warn(args...)
}

func Debug(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		return
	}
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		return
	}
	logger.Info(args...)
}
func Error(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		return
	}
	logger.Error(args...)
}

func Infof(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		return
	}
	logger.Infof(format, v...)
}
func Warnf(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		return
	}
	logger.Warnf(format, v...)
}
func Debugf(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		return
	}
	logger.Debugf(format, v...)
}
func Errorf(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		return
	}
	logger.Errorf(format, v...)

}

func GetLogger() *Logger {
	return logger
}
func (l *Logger) Println(v ...interface{}) {
	switch l.level {
	case DebugLevel:
		l.Debug(v...)
		break
	case InfoLevel:
		l.Info(v...)
		break
	case WarnLevel:
		l.Warn(v...)
		break
	case ErrorLevel:
		l.Error(v...)
		break
	default:
		fmt.Println(v...)
		break

	}
}
func (l *Logger) Printf(format string, v ...interface{}) {
	switch l.level {
	case DebugLevel:
		l.Debugf(format, v...)
		break
	case InfoLevel:
		l.Infof(format, v...)
		break
	case WarnLevel:
		l.Warnf(format, v...)
		break
	case ErrorLevel:
		l.Errorf(format, v...)
		break
	default:
		fmt.Printf(format, v...)

	}
}
