package routers

import (
	"downloader/common"
	"downloader/download"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

var fileLock sync.Mutex

// 获取任务信息
func getFileInfo(c *gin.Context) {
	result := common.NewResult()
	defer c.JSON(http.StatusOK, result)
	var url string
	err := c.BindJSON(&url)
	if err != nil {
		result.Code = -1
		result.Msg = fmt.Sprint(err)
		return
	}
	fileList, err := common.GetFileList(url)
	FileList = fileList
}

// 添加任务
func addTask(c *gin.Context) {
	result := common.NewResult()
	defer c.JSON(http.StatusOK, result)
	var fileInfo common.FileInfo
	err := c.BindJSON(&fileInfo)
	if err != nil {
		result.Code = -1
		result.Msg = fmt.Sprint(err)
		return
	}
	id, err := download.Download.AddTask(fileInfo, common.NewClient())
	if err != nil {
		result.Code = -1
		result.Msg = fmt.Sprint(err)
		return
	}
	result.Data = id
}

// 对任务列表的操作
func operate(c *gin.Context) {
	result := common.NewResult()
	defer c.JSON(http.StatusOK, result)
	var event download.DownloadEvent
	err := c.BindJSON(&event)
	if err != nil {
		result.Code = -1
		result.Msg = fmt.Sprint(err)
		return
	}
	download.Download.Event <- event
}

// UI快照
func saveUI(c *gin.Context) {
	result := common.NewResult()
	defer c.JSON(http.StatusOK, result)
	var html string
	err := c.BindJSON(&html)
	if err == nil {
		common.DBLock.Lock()
		_, err = common.DB.Exec("delete from ui")
		common.DBLock.Unlock()
		common.DBLock.Lock()
		_, err = common.UIInsert.Exec(html)
		common.DBLock.Unlock()
	}
}
