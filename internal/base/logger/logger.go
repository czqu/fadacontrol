package logger

import (
	"context"
	"errors"
	_ "fadacontrol/internal/base/conf"
	conf "fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	_log "fadacontrol/internal/base/log"
	"fadacontrol/pkg/syncer"
	"fadacontrol/pkg/sys/log"
	"fadacontrol/pkg/utils"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)
import "gopkg.in/natefinch/lumberjack.v2"

var logger *Logger

type Logger struct {
	logger          *zap.Logger
	sugar           *zap.SugaredLogger
	level           _log.Loglevel
	reportLevel     _log.Loglevel
	ctx             context.Context
	logPath         string
	logLevel        string
	logOutputSyncer *syncer.MultiBufferSyncWriteSyncer
	LogReporter     _log.LogReporter
}

var once sync.Once

func NewLogger(ctx context.Context) *Logger {
	return &Logger{ctx: ctx}
}

func (l *Logger) AddReader(reader io.Writer) int {

	return l.logOutputSyncer.AddSyncerAndFlushBuf(syncer.AddSync(reader))

}
func (l *Logger) RemoveWriter(id int) {

	l.logOutputSyncer.Remove(id)
}
func InitLog(ctx context.Context) {
	once.Do(func() {
		if logger != nil {
			return
		}
		l := new(Logger)
		logger = l
		l.logOutputSyncer = syncer.NewMultiBufferSyncWriteSyncer(100)
		var err error
		_conf := utils.GetValueFromContext(ctx, constants.ConfKey, conf.NewDefaultConf())

		logger.logPath, err = filepath.Abs(_conf.GetWorkdir() + "/log/" + _conf.LogName)
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
func InitLogReporter(options *_log.SentryOptions) {
	if logger == nil {
		return
	}
	logger.LogReporter = _log.NewSentryReporter(options)
	logger.reportLevel = str2Loglevel(strings.ToLower(options.Level))

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

func (l *Logger) Init(logPath string, loglevel _log.Loglevel) error {

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

	if !loglevel.IsValid() {

		return errors.New("invalid log level")
	}
	l.level = loglevel
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(l.level.ZapLevel())
	logHook := lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    1,
		MaxBackups: 100,
		MaxAge:     30,
		Compress:   false,
	}
	flushInterval := 1 * time.Second
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
		zapcore.NewMultiWriteSyncer(bufferedWriteSyncer, zapcore.AddSync(os.Stdout), zapcore.AddSync(l.logOutputSyncer)),
		atomicLevel,
	)

	var coreArr []zapcore.Core

	coreArr = append(coreArr, core)
	if loglevel == _log.DebugLevel {
		l.logger = zap.New(zapcore.NewTee(coreArr...), zap.AddCallerSkip(2), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		l.logger = zap.New(zapcore.NewTee(coreArr...), zap.AddCallerSkip(2), zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
	}

	l.sugar = l.logger.Sugar()
	return nil
}
func (l *Logger) ReportInfoMsg(msg string) {
	if l.reportLevel > _log.InfoLevel {
		return
	}
	if l.LogReporter != nil {
		(l.LogReporter).ReportEvent(msg, _log.InfoLevel)
	}
}
func (l *Logger) ReportWarnMsg(msg string) {
	if l.reportLevel > _log.WarnLevel {
		return
	}
	if l.LogReporter != nil {
		(l.LogReporter).ReportEvent(msg, _log.WarnLevel)
	}
}
func (l *Logger) ReportErrorMsg(msg string) {
	if l.reportLevel > _log.ErrorLevel {
		return
	}
	if l.LogReporter != nil {
		(l.LogReporter).ReportEvent(msg, _log.ErrorLevel)
	}
}
func (l *Logger) ReportFatalMsg(msg string) {
	if l.LogReporter == nil {
		l.LogReporter = _log.NewSentryReporter(_log.NewDefaultSentryOptions())
	}
	(l.LogReporter).ReportEvent(msg, _log.FatalLevel)

}
func (l *Logger) ReportException(err error) {
	if l.LogReporter != nil {
		(l.LogReporter).ReportException(err)
	}
}
func (l *Logger) Sync() {

	l.logger.Sync()
}

func (l *Logger) Debug(args ...interface{}) {

	l.sugar.Debug(args)
}
func (l *Logger) Info(args ...interface{}) {

	l.sugar.Info(args)
	l.ReportInfoMsg(fmt.Sprint(args...))
}
func (l *Logger) Warn(args ...interface{}) {
	l.sugar.Warn(args)
	l.ReportWarnMsg(fmt.Sprint(args...))
}
func (l *Logger) Error(args ...interface{}) {

	l.sugar.Error(args)
	l.logger.Sync()
	l.ReportErrorMsg(fmt.Sprint(args...))
}
func (l *Logger) Fatal(args ...interface{}) {

	l.sugar.Error(args)
	l.logger.Sync()
	l.ReportFatalMsg(fmt.Sprint(args...))
}
func (l *Logger) Debugf(format string, v ...interface{}) {

	l.sugar.Debugf(format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {

	l.sugar.Infof(format, v...)
	l.ReportInfoMsg(fmt.Sprintf(format, v...))
}
func (l *Logger) Warnf(format string, v ...interface{}) {

	l.sugar.Warnf(format, v...)
	l.ReportWarnMsg(fmt.Sprintf(format, v...))
}
func (l *Logger) Errorf(format string, v ...interface{}) {

	l.sugar.Errorf(format, v...)
	l.logger.Sync()
	l.ReportErrorMsg(fmt.Sprintf(format, v...))
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.sugar.Fatalf(format, v...)
	l.logger.Sync()
	l.ReportFatalMsg(fmt.Sprintf(format, v...))
}

var loglevel _log.Loglevel = _log.InfoLevel

func str2Loglevel(level string) _log.Loglevel {
	switch level {
	case "debug":
		loglevel = _log.DebugLevel
		break
	case "info":
		loglevel = _log.InfoLevel
		break
	case "warn":
		loglevel = _log.WarnLevel
		break
	case "error":
		loglevel = _log.ErrorLevel
		break
	case "fatal":
		loglevel = _log.FatalLevel
	default:
		loglevel = _log.InfoLevel

	}
	return loglevel
}

func Sync() {
	if logger == nil {
		return
	}
	logger.Sync()

}

func Warn(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		log.Warnf(nil, fmt.Sprintf("%v", args...))
		return
	}
	logger.Warn(args...)
}

func Debug(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		log.Debugf(nil, fmt.Sprintf("%v", args...))
		return
	}
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		log.Infof(nil, fmt.Sprintf("%v", args...))
		return
	}
	logger.Info(args...)
}
func Error(args ...interface{}) {
	if logger == nil {
		fmt.Println(args...)
		log.Errorf(nil, fmt.Sprintf("%v", args...))
		return
	}
	logger.Error(args...)
	logger.Sync()
	log.Errorf(nil, fmt.Sprintf("%v", args...))
}
func Fatal(args ...interface{}) {
	if logger == nil {
		r := _log.NewSentryReporter(_log.NewDefaultSentryOptions())
		fmt.Println(args...)
		log.Fatalf(nil, fmt.Sprintf("%v", args...))
		r.ReportEvent(fmt.Sprintf("%v", args...), _log.FatalLevel)
		return
	}
	logger.Fatal(args...)
	logger.Sync()
	log.Fatalf(nil, fmt.Sprintf("%v", args...))
}
func Infof(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		log.Infof(nil, format, v...)
		return
	}
	logger.Infof(format, v...)
}
func Warnf(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		log.Warnf(nil, format, v...)
		return
	}
	logger.Warnf(format, v...)
}
func Debugf(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		log.Debugf(nil, format, v...)
		return
	}
	logger.Debugf(format, v...)
}
func Errorf(format string, v ...interface{}) {
	if logger == nil {
		fmt.Printf(format, v...)
		log.Errorf(nil, format, v...)
		return
	}
	logger.Errorf(format, v...)
	logger.Sync()
	log.Errorf(nil, format, v...)

}

func GetLogger() *Logger {
	return logger
}
func (l *Logger) Println(v ...interface{}) {
	switch l.level {
	case _log.DebugLevel:
		l.Debug(v...)
		break
	case _log.InfoLevel:
		l.Info(v...)
		break
	case _log.WarnLevel:
		l.Warn(v...)
		break
	case _log.ErrorLevel:
		l.Error(v...)
		break
	case _log.FatalLevel:
		l.Fatal(v...)
	default:
		fmt.Println(v...)
		break

	}
}
func (l *Logger) Printf(format string, v ...interface{}) {
	switch l.level {
	case _log.DebugLevel:
		l.Debugf(format, v...)
		break
	case _log.InfoLevel:
		l.Infof(format, v...)
		break
	case _log.WarnLevel:
		l.Warnf(format, v...)
		break
	case _log.ErrorLevel:
		l.Errorf(format, v...)
		break
	case _log.FatalLevel:
		l.Fatalf(format, v...)
	default:
		fmt.Printf(format, v...)

	}
}
