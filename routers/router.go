package routers

import (
	"downloader/common"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	Conn     *websocket.Conn
	SetConn  *websocket.Conn
	FileConn *websocket.Conn
	FileList []common.FileInfo
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	router.Use(gin.Logger())

	router.Use(gin.Recovery())

	gin.SetMode(common.RunMode)

	// websocket
	router.GET("/getSetting", setting)

	router.GET("/getTaskInfo", taskInfo)

	router.GET("/main", mainUI)

	router.GET("/fileList", fileList)

	// http
	router.POST("/getFileInfo", getFileInfo)

	router.POST("/addTask", addTask)

	router.POST("/operate", operate)

	router.POST("/saveUI", saveUI)

	return router
}
