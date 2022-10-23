package main

import (
	"context"
	"fmt"
	"iris-demo/web"
	"time"

	"github.com/kataras/iris/v12"
)

func main() {
	app := iris.Default()

	//响应记录
	// start record.
	app.Use(func(ctx iris.Context) {
		ctx.Record()
		ctx.Next()
	})
	// collect and "log".
	app.Done(func(ctx iris.Context) {
		body := ctx.Recorder().Body()
		// Should print success.
		app.Logger().Infof("%s sent: %s", ctx.Request().URL.RequestURI(), string(body))
	})

	//优雅关闭
	iris.RegisterOnInterrupt(func() {
		timeout := 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		fmt.Println("app closed")

		// close all hosts
		app.Shutdown(ctx)
	})
	//注册全局中间件
	app.Use(globalMiddleware)

	//设置路由规则，强制路由handlerx之星完成后执行ctx.Next()
	app.SetExecutionRules(iris.ExecutionRules{
		Done: iris.ExecutionOptions{Force: true},
	})

	//注册业务路由
	web.RegisterRoute(app)

	//生成iris配置, 并通过配置启动
	config := iris.WithConfiguration(iris.Configuration{
		DisableStartupLog: false,
		Charset:           "UTF-8",
	})
	app.Run(iris.Addr(":8088"), config)
}

// 自定义中间件
func globalMiddleware(ctx iris.Context) {
	ctx.Application().Logger().Infof("globalMiddleware -> Runs before: %s", ctx.Path())
	ctx.Next()
}
