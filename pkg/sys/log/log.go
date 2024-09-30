package log

import "runtime"

type SysLogInterface interface {
	Debugf(error error, format string, v ...interface{})
	Infof(error error, format string, v ...interface{})
	Warnf(error error, format string, v ...interface{})
	Errorf(error error, format string, v ...interface{})
	Fatalf(error error, format string, v ...interface{})
	Close() error
	SetSkip(skip int)
}

var log SysLogInterface

func GetLogSys() SysLogInterface {
	return log

}
func Debugf(error error, format string, v ...interface{}) {
	if log == nil {
		return
	}
	log.Debugf(error, format, v...)
}
func Infof(error error, format string, v ...interface{}) {
	if log == nil {
		return
	}
	log.Infof(error, format, v...)
}
func Warnf(error error, format string, v ...interface{}) {
	if log == nil {
		return
	}
	log.Warnf(error, format, v...)
}
func Errorf(error error, format string, v ...interface{}) {
	if log == nil {
		return
	}
	log.Errorf(error, format, v...)
}
func Fatalf(error error, format string, v ...interface{}) {
	if log == nil {
		return
	}
	log.Fatalf(error, format, v...)

}
func GetFileLine(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	return file, line
}
