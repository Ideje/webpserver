package app

import (
	"context"
	"fmt"
	"github.com/inconshreveable/log15"
	"math/rand"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	Enabled bool
	LogFile string
	Color   = true

	OutHandler log15.Handler
	LogLevel   log15.Lvl

	Session int64

	Log = log15.New()
)

type Ctx struct {
	UserData   interface{}
	Session    int64
	Time_start time.Time
	Ctx        *context.Context
}

func NewCtx() *Ctx {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return &Ctx{
		Session:    atomic.AddInt64(&Session, 1),
		Time_start: time.Now(),
		Ctx:        &ctx,
	}
}

func (c *Ctx) prependCtxSession(ctx []interface{}) []interface{} {
	return append([]interface{}{"ctx", c.Session}, ctx...)
}

func (c *Ctx) Debug(msg string, ctx ...interface{}) {
	ctx = c.prependCtxSession(ctx)
	Log.Debug(msg, ctx...)
}

func (c *Ctx) Info(msg string, ctx ...interface{}) {
	ctx = c.prependCtxSession(ctx)
	Log.Info(msg, ctx...)
}

func (c *Ctx) Warn(msg string, ctx ...interface{}) {
	ctx = c.prependCtxSession(ctx)
	Log.Warn(msg, ctx...)
}

func (c *Ctx) Error(msg string, ctx ...interface{}) {
	ctx = c.prependCtxSession(ctx)
	Log.Error(msg, ctx...)
}

func init() {
	Log.SetHandler(log15.DiscardHandler())

	rand.Seed(time.Now().UTC().UnixNano())

	var logfmt string
	if Color {
		OutHandler = log15.StreamHandler(os.Stdout, log15.TerminalFormat())
		logfmt = "terminal"
	} else {
		OutHandler = log15.StreamHandler(os.Stdout, log15.LogfmtFormat())
		logfmt = "logfmt"
	}

	SetLogLevel("debug")
	Log.Debug("Log format", "color", Color, "type", logfmt)

}

func SetLogLevel(lvlString string) {
	if logLevel, err := log15.LvlFromString(lvlString); err == nil {
		LogLevel = logLevel
		Log.SetHandler(log15.DiscardHandler())
		Log.SetHandler(log15.MultiHandler(
			log15.LvlFilterHandler(LogLevel, log15.CallerFileHandler(OutHandler)),
		))
		Log.Info("Set log level", "level", LogLevel)
	}
}

func LogError(err error) (b bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log where
		// the error happened, 0 = this function, we don't want that.
		_, fn, line, _ := runtime.Caller(1)
		//log.Printf("[error] %s:%d %v", fn, line, err)
		Log.Warn("error", "err", err, "file", fmt.Sprintf(" %s:%d ", fn, line))
		b = true
	}
	return
}
