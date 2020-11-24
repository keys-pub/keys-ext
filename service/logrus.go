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

func setupLogrus(logLevel LogLevel, logPath string) (*logrus.Logger, logrusInterceptor) {
	llog := logrus.StandardLogger()
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	llog.SetLevel(logrusFromLevel(logLevel))
	llog.SetFormatter(formatter)
	// slog.SetReportCaller(true)

	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, filePerms) // #nosec
		if err != nil {
			logFatal(err)
		}
		llog.Infof("Logging to %s", logPath)
		llog.SetOutput(logFile)
	}
	return llog, newLogrusInterceptor(llog)
}

// newPackageLogger adds package name to logrus.Logger.
func newPackageLogger(log *logrus.Logger, packageName string) Logger {
	return &packageLogger{log, packageName}
}

type packageLogger struct {
	*logrus.Logger
	packageName string
}

func (l packageLogger) Debugf(format string, args ...interface{}) {
	l.WithField("package", l.packageName).Debugf(format, args...)
}

func (l packageLogger) Infof(format string, args ...interface{}) {
	l.WithField("package", l.packageName).Infof(format, args...)
}

func (l packageLogger) Warningf(format string, args ...interface{}) {
	l.WithField("package", l.packageName).Warningf(format, args...)
}

func (l packageLogger) Errorf(format string, args ...interface{}) {
	l.WithField("package", l.packageName).Errorf(format, args...)
}

func (l packageLogger) Fatalf(format string, args ...interface{}) {
	l.WithField("package", l.packageName).Fatalf(format, args...)
}
