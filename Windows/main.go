package main

import (
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"

	"github.com/UserExistsError/conpty"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// 嵌入www目录
var (
	//go:embed www/*
	webDir embed.FS
)

func main() {
	// 接收一个传入的参数port，默认2223
	port := ":2223"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	gin.SetMode(gin.ReleaseMode)

	var c *exec.Cmd
	c = exec.Command("cmd")

	var f interface{}
	var err error
	// 使用ConPTY启动Windows终端
	f, err = conpty.Start(c.Path)
	if err != nil {
		panic(err)
	}

	m := melody.New()

	go func() {
		for {
			buf := make([]byte, 1024)
			var read int
			// 使用ConPTY的读取方法
			read, err = f.(*conpty.ConPty).Read(buf)
			if err != nil {
				break
			}
			m.Broadcast(buf[:read])
		}
	}()

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		// 使用ConPTY的写入方法
		f.(*conpty.ConPty).Write(msg)
	})

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	staticFp, _ := fs.Sub(webDir, "www")
	r.NoRoute(gin.WrapH(http.FileServer(http.FS(staticFp))))
	fmt.Println("Server running on port " + port)
	r.Run(port)
}
