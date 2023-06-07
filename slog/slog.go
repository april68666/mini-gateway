package slog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
)

var LevelMap = map[int32]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var defaultSlog *Slog

func init() {
	defaultSlog = NewSlog(os.Stdout, 1024)
	defaultSlog.SetCallDepth(2)
}

func InitSlog(log *Slog) {
	defaultSlog = log
}

func NewSlog(out io.Writer, bufLen int) *Slog {
	s := &Slog{
		Logger:    log.New(out, "", 0),
		buff:      make(chan *logInfo, bufLen),
		flush:     make(chan *sync.WaitGroup),
		callDepth: 2,
	}
	go s.writeLog()
	return s
}

type logInfo struct {
	Level    int32
	Format   string
	V        []any
	Location string
	Time     time.Time
}

type Slog struct {
	*log.Logger
	level     int32
	buff      chan *logInfo
	flush     chan *sync.WaitGroup
	callDepth int
}

func (s *Slog) SetCallDepth(depth int) {
	s.callDepth = depth
}

func (s *Slog) SetLevel(level int) {
	atomic.StoreInt32(&s.level, int32(level))
}

func (s *Slog) Debug(format string, v ...any) {
	s.writeBuf(DEBUG, format, v...)
}

func (s *Slog) Info(format string, v ...any) {
	s.writeBuf(INFO, format, v...)
}

func (s *Slog) Warn(format string, v ...any) {
	s.writeBuf(WARN, format, v...)
}

func (s *Slog) Error(format string, v ...any) {
	s.writeBuf(ERROR, format, v...)
}

func (s *Slog) Fatal(format string, v ...any) {
	s.writeBuf(FATAL, format, v...)
}

func (s *Slog) writeBuf(level int32, format string, v ...any) {
	s.buff <- &logInfo{
		Level:    level,
		Format:   format,
		V:        v,
		Location: s.getLocation(s.callDepth + 2),
		Time:     time.Now(),
	}
}

func (s *Slog) writeLog() {
	for {
		select {
		case msg := <-s.buff:
			s.fmtPrint(msg)
		case wg := <-s.flush:
			s.writeAllBuf()
			wg.Done()

		}
	}
}

func (s *Slog) writeAllBuf() {
	isBreak := false
	for !isBreak {
		select {
		case msg := <-s.buff:
			s.fmtPrint(msg)
		default:
			isBreak = true
		}
	}
}

func (s *Slog) fmtPrint(info *logInfo) {
	if s.level <= atomic.LoadInt32(&info.Level) {
		msg := fmt.Sprintf("%s %s %s %s", info.Time.Format("2006-01-02 15:04:05,612"), LevelMap[info.Level], info.Location, info.Format)
		if len(info.V) > 0 {
			s.Printf(msg, info.V...)
		} else {
			s.Println(msg)
		}
	}
}

func (s *Slog) getLocation(skip int) string {
	_, fine, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s:%d", fine, line)
}

// Flush 等待所有日志写入完成，成功返回 true 超时返回 false
func (s *Slog) Flush(ctx context.Context) bool {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	select {
	case s.flush <- wg:
		wg.Wait()
		return true
	case <-ctx.Done():
		return false
	}
}

func Debug(format string, v ...any) {
	defaultSlog.Debug(format, v...)
}

func Info(format string, v ...any) {
	defaultSlog.Info(format, v...)
}

func Warn(format string, v ...any) {
	defaultSlog.Warn(format, v...)
}

func Error(format string, v ...any) {
	defaultSlog.Error(format, v...)
}

func Fatal(format string, v ...any) {
	defaultSlog.Fatal(format, v...)
}

func Flush(ctx context.Context) bool {
	return defaultSlog.Flush(ctx)
}
