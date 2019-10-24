package routers

import (
	"downloader/conf"
	"downloader/downloader"
	"downloader/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	Conn    *websocket.Conn
	SetConn *websocket.Conn
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	router.Use(gin.Logger())

	router.Use(gin.Recovery())

	gin.SetMode(conf.RunMode)

	router.GET("/getSetting", func(c *gin.Context) {
		conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(
			c.Writer, c.Request, nil)
		if err != nil {
			http.NotFound(c.Writer, c.Request)
			return
		}
		SetConn = conn
		_ = conn.WriteJSON(downloader.Download)
		var operate struct {
			Op            int    `json:"op"`
			SavePath      string `json:"savePath"`
			MaxRoutineNum int    `json:"maxRoutineNum"`
		}
		for {
			err := conn.ReadJSON(&operate)
			if err == nil {
				_ = Conn.WriteJSON(operate)
				if operate.Op == 5 {
					downloader.Download.SavePath = operate.SavePath
					downloader.Download.MaxRoutineNum = operate.MaxRoutineNum
					conf.SetValue(operate.SavePath, operate.MaxRoutineNum, 3)
				}
				continue
			}
			if _, ok := err.(*net.OpError); ok || websocket.IsCloseError(err, 1001, 1006) {
				return
			}
		}
	})

	router.GET("/getFileInfo", func(c *gin.Context) {
		result := util.NewResult()
		defer c.JSON(http.StatusOK, result)
		fileInfo, err := util.GetFileInfo(c.Query("url"), util.NewClient())
		if err != nil {
			result.Code = -1
			result.Msg = fmt.Sprint(err)
			return
		}
		fileInfo.SavePath = downloader.Download.SavePath
		result.Data = fileInfo
	})

	router.POST("/addTask", func(c *gin.Context) {
		result := util.NewResult()
		defer c.JSON(http.StatusOK, result)
		var fileInfo util.FileInfo
		err := c.BindJSON(&fileInfo)
		if err != nil {
			result.Code = -1
			result.Msg = fmt.Sprint(err)
			return
		}
		id, err := downloader.Download.AddTask(fileInfo, util.NewClient())
		if err != nil {
			result.Code = -1
			result.Msg = fmt.Sprint(err)
			return
		}
		result.Data = id
	})

	router.POST("/operate", func(context *gin.Context) {
		result := util.NewResult()
		defer context.JSON(http.StatusOK, result)
		var event downloader.DownloadEvent
		err := context.BindJSON(&event)
		if err != nil {
			result.Code = -1
			result.Msg = fmt.Sprint(err)
			return
		}
		downloader.Download.Event <- event
	})

	router.GET("/getTaskInfo", func(c *gin.Context) {
		// change the reqest to websocket model
		conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(
			c.Writer, c.Request, nil)
		if err != nil {
			http.NotFound(c.Writer, c.Request)
			return
		}
		task, ok := downloader.Download.ActiveTaskMap[c.Query("id")]
		if !ok {
			_ = conn.Close()
			return
		}
		task.Conn = conn
	})

	router.GET("/checkActive", func(c *gin.Context) {
		// change the reqest to websocket model
		conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(
			c.Writer, c.Request, nil)
		if err != nil {
			http.NotFound(c.Writer, c.Request)
			return
		}
		Conn = conn
		Conn.SetCloseHandler(func(code int, text string) error {
			_ = conf.DB.Close()
			downloader.Download.BeforeExit()
			time.Sleep(time.Second)
			log.Fatal("主动断开链接")
			return nil
		})
		var operate struct {
			Op            int    `json:"op"`
			SavePath      string `json:"savePath"`
			MaxRoutineNum int    `json:"maxRoutineNum"`
		}
		for {
			err := Conn.ReadJSON(&operate)
			if err == nil {
				_ = SetConn.WriteJSON(operate)
				continue
			}
			if _, ok := err.(*net.OpError); ok || websocket.IsCloseError(err, 1001, 1006) {
				downloader.Download.BeforeExit()
				time.Sleep(time.Second)
				log.Fatal("断开链接")
			}
		}
	})

	return router
}
