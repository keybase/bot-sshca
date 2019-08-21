package kssh

import log "github.com/sirupsen/logrus"

// Prefix formatter adds a prefix of "kssh: " to log messages before delegating to the default text formatter
type prefixFormatter struct{}

func (pf *prefixFormatter) Format(entry *log.Entry) ([]byte, error) {
	entry.Message = "kssh: " + entry.Message
	textFormatter := &log.TextFormatter{}
	return textFormatter.Format(entry)
}

func InitLogging() {
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&prefixFormatter{})
}
