package routers

import (
	"downloader/common"
	"downloader/downloader"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type operation struct {
	Op            int               `json:"op"`
	SavePath      string            `json:"savePath"`
	MaxRoutineNum int               `json:"maxRoutineNum"`
	FileList      []common.FileInfo `json:"fileList"`
}

// 与主界面的信息交互
func mainUI(c *gin.Context) {
	// change the reqest to websocket model
	conn, err := upgrade(c)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	Conn = conn
	Conn.SetCloseHandler(func(code int, text string) error {
		_ = common.DB.Close()
		downloader.Download.BeforeExit()
		time.Sleep(time.Second)
		log.Fatal("主动断开链接")
		return nil
	})
	file, err := os.OpenFile("data/UI.txt", os.O_RDONLY, 0644)
	if err == nil {
		content, err := ioutil.ReadAll(file)
		if err == nil {
			_ = Conn.WriteJSON(string(content))
		}
	}
	var operate operation
	for {
		err := Conn.ReadJSON(&operate)
		if err == nil {
			_ = SetConn.WriteJSON(operate)
			continue
		}
		if exit(err) {
			downloader.Download.BeforeExit()
			time.Sleep(time.Second)
			log.Fatal("断开链接")
		}
	}
}

// 与设置界面的信息交互
func setting(c *gin.Context) {
	conn, err := upgrade(c)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	SetConn = conn
	_ = conn.WriteJSON(downloader.Download)
	var operate operation
	for {
		err := conn.ReadJSON(&operate)
		if err == nil {
			_ = Conn.WriteJSON(operate)
			if operate.Op == 5 {
				downloader.Download.SavePath = operate.SavePath
				downloader.Download.MaxRoutineNum = operate.MaxRoutineNum
				common.SetValue(operate.SavePath, operate.MaxRoutineNum, 3)
			}
			continue
		}
		if exit(err) {
			return
		}
	}
}

// 与下载信息列表界面的交互
func fileList(c *gin.Context) {
	conn, err := upgrade(c)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	FileConn = conn
	_ = conn.WriteJSON(FileList)
	var operate operation
	for {
		err := conn.ReadJSON(&operate)
		if err == nil {
			_ = Conn.WriteJSON(operate)
			continue
		}
		if exit(err) {
			return
		}
	}
}

// 与单个任务的信息交互
func taskInfo(c *gin.Context) {
	// change the reqest to websocket model
	conn, err := upgrade(c)
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
}

func upgrade(c *gin.Context) (*websocket.Conn, error) {
	return (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(
		c.Writer, c.Request, nil)
}

func exit(err error) bool {
	if _, ok := err.(*net.OpError); ok || websocket.IsCloseError(err, 1001, 1006) {
		return true
	}
	return false
}
