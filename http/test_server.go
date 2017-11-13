package client

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
	"time"
)

type User struct {
	Name string
	Age int
	Address string
	City string
}

var bob *User = &User{Name: "Bob", Age: 34, Address: "Street 11 B", City: "Townsville"}

func startTestServer() *iris.Application {
	app := iris.New()
	app.Get("/test/ping", func(ctx iris.Context) {
		ctx.Text("pong")
	})
	app.Get("/test/v1/users/bob", func(ctx iris.Context) {
		ctx.JSON(bob)
	})
	app.Delete("/test/v1/users/bob", func(ctx iris.Context) {
		ctx.JSON(bob)
	})
	app.Post("/test/v1/users/bob", func(ctx iris.Context) {
		ctx.JSON(bob)
	})
	app.Put("/test/v1/users/bob", func(ctx iris.Context) {
		ctx.JSON(bob)
	})

	ws := websocket.New(websocket.Config{
		BinaryMessages: true,
		ReadBufferSize:  100*1024,
		WriteBufferSize: 100*1024,
		MaxMessageSize: 20*1024,
	})

	app.Get("/ws", ws.Handler())

	ws.OnConnection(func(c websocket.Connection) {
		c.OnMessage(func(msg []byte) {
			c.EmitMessage(msg)
		})
		c.OnDisconnect(func() {
		})
	})
	go app.Run(iris.Addr(":9000"))

	time.Sleep(100 * time.Millisecond)
	return app
}