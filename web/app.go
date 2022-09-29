package web

import (
	"fmt"

	"github.com/kataras/iris/v12"
)

const maxSize = 8 * iris.MB

func RegisterRoute(app *iris.Application) {
	rootPath := app.Party("/")
	rootPath.Use(func(ctx iris.Context) {
		fmt.Printf("/ request\n")
	})

	app.Handle("GET", "/ping", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"message": "pong"})
	})

	app.Get("/user/{id:uint64}", func(ctx iris.Context) {
		userID, _ := ctx.Params().GetUint64("id")
		ctx.Writef("User ID: %d", userID)
	})

	path1 := app.Party("/path1")
	path1.UseRouter(func(ctx iris.Context) {
		fmt.Printf("path1 request\n")
		ctx.Next()
	})

	path1.Post("/post", func(ctx iris.Context) {
		ids := ctx.URLParamSlice("id")
		id, err := ctx.URLParamInt("id")
		if err != nil {
			ctx.StopWithError(iris.StatusBadRequest, err)
			return
		}

		age := ctx.URLParamIntDefault("age", 0)
		name := ctx.PostValue("name")
		message := ctx.PostValue("message")

		ctx.Writef("ids: %v; id: %d; age: %d; name: %s; message: %s", ids, id, age, name, message)
	})

	path1.Post("/uploads", func(ctx iris.Context) {
		ctx.SetMaxRequestBodySize(maxSize)

		_, fileHeader, err := ctx.FormFile("file")
		if err != nil {
			ctx.StopWithError(iris.StatusBadRequest, err)
			return
		}

		// Upload the file to specific destination.
		// dest := filepath.Join("./uploads", fileHeader.Filename)
		// ctx.SaveFormFile(fileHeader, dest)
		ctx.SaveFormFile(fileHeader, "uploads/"+fileHeader.Filename)
		ctx.Writef("File: %s uploaded!", fileHeader.Filename)
	})
}
