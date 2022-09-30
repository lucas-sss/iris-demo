package web

import (
	"fmt"
	"strings"

	"github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
)

const maxSize = 8 * iris.MB

func RegisterRoute(app *iris.Application) {

	//使用jwt
	j := jwt.New(jwt.Config{
		// Extract by "token" url parameter.
		Extractor: func(ctx iris.Context) (string, error) {
			authHeader := ctx.GetHeader("Authorization")
			if authHeader == "" {
				return "", nil // No error, just no token
			}
			// TODO: Make this a bit more robust, parsing-wise
			authHeaderParts := strings.Split(authHeader, " ")
			if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
				return "", fmt.Errorf("authorization header format must be Bearer {token}")
			}
			return authHeaderParts[1], nil
		},

		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("My Secret"), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	rootPath := app.Party("/")
	rootPath.Use(j.Serve)

	app.Get("/login", func(ctx iris.Context) {
		token := jwt.NewTokenWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"foo": "bar",
		})
		// Sign and get the complete encoded token as a string using the secret
		tokenString, _ := token.SignedString([]byte("My Secret"))
		ctx.JSON(iris.Map{"token": tokenString})
	})

	app.Handle("GET", "/ping", func(ctx iris.Context) {
		ctx.Application().Logger().Infof("ping request")
		ctx.JSON(iris.Map{"message": "pong"})
	})

	app.Get("/user/{id:uint64}", func(ctx iris.Context) {
		userID, _ := ctx.Params().GetUint64("id")
		ctx.Writef("User ID: %d", userID)
	})

	path1 := app.Party("/path1")
	path1.UseRouter(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("path1 request")
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
