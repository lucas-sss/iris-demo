package web

import (
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/v12"
)

const maxSize = 8 * iris.MB

func authenticatedHandler(ctx iris.Context) {
	user := ctx.Values().Get("jwt").(*jwt.Token)

	foobar := user.Claims.(jwt.MapClaims)
	for key, value := range foobar {
		fmt.Printf("Jwt map info %s = %s\n", key, value)
	}
	ctx.Next()

	// ctx.Writef("This is an authenticated request\n")
	// ctx.Writef("Claim content:\n")

	// foobar := user.Claims.(jwt.MapClaims)
	// for key, value := range foobar {
	// 	ctx.Writef("%s = %s", key, value)
	// }
}

func RegisterRoute(app *iris.Application) {

	//使用jwt
	jwtMiddleware := jwt.New(jwt.Config{
		// 注意，新增了一个错误处理函数
		ErrorHandler: func(ctx iris.Context, err error) {
			if err == nil {
				return
			}

			ctx.StopExecution()
			ctx.StatusCode(iris.StatusOK)
			// ctx.StatusCode(iris.StatusUnauthorized)
			ctx.JSON(iris.Map{
				"Code": "401",
				"Msg":  err.Error(),
			})
		},
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

	app.Get("/login", func(ctx iris.Context) {
		token := jwt.NewTokenWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"foo": "bar",

			// 签发时间
			"iat": time.Now().Unix(),
			// 设定过期时间，便于测试，设置1分钟过期
			"exp": time.Now().Add(1 * time.Minute * time.Duration(1)).Unix(),
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
	// path1.Use(jwtMiddleware.Serve, authenticatedHandler)
	path1.Use(jwtMiddleware.Serve)
	path1.Use(authenticatedHandler)

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

	path1.Post("/upload/1", func(ctx iris.Context) {
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

	path1.Post("/upload/2", func(ctx iris.Context) {
		ctx.SetMaxRequestBodySize(maxSize)

		_, fileHeader, err := ctx.FormFile("file")
		if err != nil {
			ctx.StopWithError(iris.StatusBadRequest, err)
			return
		}

		ctx.UploadFormFiles("./uploads", beforeSave)

		ctx.Writef("File: %s uploaded!", fileHeader.Filename)
	})

}

func beforeSave(ctx iris.Context, file *multipart.FileHeader) bool {
	ip := ctx.RemoteAddr()
	// make sure you format the ip in a way
	// that can be used for a file name (simple case):
	ip = strings.Replace(ip, ".", "_", -1)
	ip = strings.Replace(ip, ":", "_", -1)

	// you can use the time.Now, to prefix or suffix the files
	// based on the current time as well, as an exercise.
	// i.e unixTime :=    time.Now().Unix()
	// prefix the Filename with the $IP-
	// no need for more actions, internal uploader will use this
	// name to save the file into the "./uploads" folder.
	file.Filename = ip + "-" + file.Filename
	return true
}
