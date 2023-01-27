package utilgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"k8s.io/klog/v2"
)

type logger struct {
	log                   klog.Logger
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
}

func NewLogger(log klog.Logger) *logger {
	return &logger{
		log:                   log,
		SkipErrRecordNotFound: true,
	}
}

func (l *logger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *logger) Info(ctx context.Context, s string, args ...interface{}) {
	l.log.V(2).Info(s, args...)
}

func (l *logger) Warn(ctx context.Context, s string, args ...interface{}) {
	l.log.V(2).Info(s, args...)
}

func (l *logger) Error(ctx context.Context, s string, args ...interface{}) {
	l.log.Error(fmt.Errorf(s), s, args...)
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	log := l.log

	sql, _ := fc()
	if l.SourceField != "" {
		log = log.WithValues(l.SourceField, utils.FileWithLineNum())
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		log.Error(err, fmt.Sprintf("[%s] %s", sql, elapsed))
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		log.V(8).Info(fmt.Sprintf("[%s] %s", sql, elapsed))
		return
	}

	log.V(8).Info(fmt.Sprintf("[%s] %s", sql, elapsed))

}
