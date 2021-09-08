package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"time"
	"webpserver/app"
	"webpserver/config"
	"webpserver/server"
)

func init() {
	err := config.LoadConfig(".")
	app.Log.Info("import config")
	if err != nil {
		app.Log.Error("cannot load config:", "err", err)
		os.Exit(1)
	}

	app.Log.Info("Config",
		"ServerAddress", config.Cfg.ServerAddress,
		"ServerName", config.Cfg.ServerName,
		"ImageSrcDir", config.Cfg.ImageSrcDir,
		"ImageDstDir", config.Cfg.ImageDstDir,
		"LogLevel", config.Cfg.LogLevel,
		"WebPMaxAge", config.Cfg.WebPMaxAge,
		"HeaderCacheInfo", config.Cfg.HeaderCacheInfo,
	)

	app.SetLogLevel(config.Cfg.LogLevel)
}

func main() {

	server := server.NewHTTPServer()

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		app.Log.Info("Go", "version", runtime.Version())
		app.Log.Info("Server started", "address", config.Cfg.ServerAddress, "cpu", runtime.NumCPU())
		if err := server.ListenAndServe(); err != nil {
			app.Log.Error("Server start", "err", err)
		}
	}()

	// shutdown
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	wait := time.Second * 3
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	app.Log.Info("Server stopping ..")
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.

	app.Log.Info("Server stopped")
	os.Exit(0)
}
