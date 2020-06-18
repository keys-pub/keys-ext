package service

import (
	"os"
	"time"

	glogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// LogInterceptor for gRPC.
type LogInterceptor interface {
	Replace()
	Unary() grpc.UnaryServerInterceptor
	Stream() grpc.StreamServerInterceptor
}

type logrusInterceptor struct {
	entry *logrus.Entry
	opts  []glogrus.Option
}

// NewLogrusInterceptor is gRPC interceptor for logrus.
func NewLogrusInterceptor(l *logrus.Logger) LogInterceptor {
	return newLogrusInterceptor(l)
}

func newLogrusInterceptor(l *logrus.Logger) logrusInterceptor {
	return logrusInterceptor{
		entry: logrus.NewEntry(l),
		opts:  []glogrus.Option{},
	}
}

func (l logrusInterceptor) Replace() {
	glogrus.ReplaceGrpcLogger(l.entry)
}

func (l logrusInterceptor) Unary() grpc.UnaryServerInterceptor {
	return glogrus.UnaryServerInterceptor(l.entry, l.opts...)
}

func (l logrusInterceptor) Stream() grpc.StreamServerInterceptor {
	return glogrus.StreamServerInterceptor(l.entry, l.opts...)
}

func setupLogrus(cfg *Config, logPath string) (*logrus.Logger, logrusInterceptor) {
	llog := logrus.StandardLogger()
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	llog.SetLevel(logrusFromLevel(cfg.LogLevel()))
	llog.SetFormatter(formatter)
	// slog.SetReportCaller(true)

	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600) // #nosec
		if err != nil {
			logFatal(err)
		}
		llog.Infof("Logging to %s", logPath)
		llog.SetOutput(logFile)
	}
	return llog, newLogrusInterceptor(llog)
}
