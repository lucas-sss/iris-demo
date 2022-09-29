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
	app.Use(myMiddleware)

	web.RegisterRoute(app)

	config := iris.WithConfiguration(iris.Configuration{
		DisableStartupLog: false,
		Charset:           "UTF-8",
	})

	app.Run(iris.Addr(":8088"), config)
	// app.Listen(":8088")
}

// 自定义中间件
func myMiddleware(ctx iris.Context) {
	ctx.Application().Logger().Infof("Runs before %s", ctx.Path())
	ctx.Next()
}
