package main

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/logger/accesslog"
	hertzzap "github.com/hertz-contrib/logger/zap"
	"github.com/hertz-contrib/pprof"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	h := server.Default(
		server.WithHostPorts("127.0.0.1:9090"),
		server.WithMaxRequestBodySize(20<<20),
		server.WithTransport(standard.NewTransporter),
	)

	pprof.Register(h)
	registerLog()

	h.Use(accesslog.New(accesslog.WithFormat("[${time}] ${status} - ${latency} ${method} ${path} ${queryParams}")))
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	h.Spin()
}

func registerLog() {
	// Customizable output directory.
	var logFilePath string
	dir := "./hlog"
	logFilePath = dir + "/logs/"
	if err := os.MkdirAll(logFilePath, 0o777); err != nil {
		log.Println(err.Error())
		return
	}

	// Set filename to date
	logFileName := time.Now().Format("2006-01-02") + ".log"
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			log.Println(err.Error())
			return
		}
	}

	// For zap detailed settings, please refer to https://github.com/hertz-contrib/logger/tree/main/zap and https://github.com/uber-go/zap
	// hlog will warp a layer of zap, so you need to calculate the depth of the caller file separately.
	logger := hertzzap.NewLogger(hertzzap.WithZapOptions(zap.AddCaller(), zap.AddCallerSkip(3)))
	// Provides compression and deletion
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    20,   // A file can be up to 20M.
		MaxBackups: 5,    // Save up to 5 files at the same time.
		MaxAge:     10,   // A file can exist for a maximum of 10 days.
		Compress:   true, // Compress with gzip.
	}
	// if you want to output the log to the file and the stdout at the same time, you can use the following codes
	fileWriter := io.MultiWriter(lumberjackLogger, os.Stdout)
	logger.SetOutput(fileWriter)
	logger.SetLevel(hlog.LevelDebug)
	hlog.SetLogger(logger)
}
