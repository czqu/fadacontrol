package log

import (
	"fadacontrol/internal/base/version"
	"github.com/getsentry/sentry-go"
	"os"
	"sync"
	"time"
)

type LogReporter interface {
	ReportMsg(msg string)
	ReportEvent(msg string, level Loglevel)
	ReportException(err error)
	Flush()
}
type SentryReporter struct {
	userId string
}

var sentryInitLock sync.Mutex

type SentryOptions struct {
	Enable             bool
	TracesSampleRate   float64
	ProfilesSampleRate float64
	UserId             string
	Level              string
}

func NewDefaultSentryOptions() *SentryOptions {
	return &SentryOptions{UserId: "unknown", TracesSampleRate: 0.2, ProfilesSampleRate: 0.2, Level: "fatal"}
}

func NewSentryReporter(options *SentryOptions) *SentryReporter {
	sentryInitLock.Lock()
	defer sentryInitLock.Unlock()
	ss := &SentryReporter{userId: options.UserId}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:                "https://82431285059e21675920c08d0e172643@o4508488989605888.ingest.us.sentry.io/4508489034825728",
		Debug:              false,
		EnableTracing:      true,
		TracesSampleRate:   options.TracesSampleRate,
		ProfilesSampleRate: options.ProfilesSampleRate,
	})
	defer sentry.Flush(2 * time.Second)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	if options.UserId == "" {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("app_info", version.GetBuildInfo())
			scope.SetTag("hostname", hostname)
		})
	} else {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{ID: ss.userId})
			scope.SetTag("app_info", version.GetBuildInfo())
			scope.SetTag("hostname", hostname)
		})
	}
	return ss
}
func LoglevelToSentryLevel(level Loglevel) sentry.Level {
	switch level {
	case DebugLevel:
		return sentry.LevelDebug
	case InfoLevel:
		return sentry.LevelInfo
	case WarnLevel:
		return sentry.LevelWarning
	case ErrorLevel:
		return sentry.LevelError
	case FatalLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

func (s *SentryReporter) ReportMsg(msg string) {
	sentry.CaptureMessage(msg)
}
func (s *SentryReporter) ReportException(err error) {
	sentry.CaptureException(err)
}
func (s *SentryReporter) ReportEvent(msg string, level Loglevel) {
	event := sentry.NewEvent()
	event.Level = LoglevelToSentryLevel(level)
	event.Message = msg
	if level == FatalLevel || level == ErrorLevel {
		event.Exception = []sentry.Exception{{
			Value: msg,
			Type:  msg,
		}}
		event.Threads = []sentry.Thread{{
			Stacktrace: sentry.NewStacktrace(),
			Current:    true,
		}}
	}
	sentry.CaptureEvent(event)
}
func (s *SentryReporter) Flush() {

}
