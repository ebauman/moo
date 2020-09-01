package logger

import log "github.com/sirupsen/logrus"

func BuildLogger(level string) *log.Logger {
	l := log.New()

	switch level {
	case "trace":
		l.SetLevel(log.TraceLevel)
		break
	case "debug":
		l.SetLevel(log.DebugLevel)
		break
	case "info":
		l.SetLevel(log.InfoLevel)
		break
	case "warning":
		l.SetLevel(log.WarnLevel)
		break
	case "error":
		l.SetLevel(log.ErrorLevel)
		break
	case "fatal":
		l.SetLevel(log.FatalLevel)
		break
	case "panic":
		l.SetLevel(log.PanicLevel)
		break
	default:
		l.SetLevel(log.InfoLevel)
	}

	return l
}