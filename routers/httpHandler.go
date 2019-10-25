package routers

import (
	"downloader/common"
	"downloader/downloader"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
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
	id, err := downloader.Download.AddTask(fileInfo, common.NewClient())
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
	var event downloader.DownloadEvent
	err := c.BindJSON(&event)
	if err != nil {
		result.Code = -1
		result.Msg = fmt.Sprint(err)
		return
	}
	downloader.Download.Event <- event
}

// UI快照
func saveUI(c *gin.Context) {
	result := common.NewResult()
	defer c.JSON(http.StatusOK, result)
	var html string
	err := c.BindJSON(&html)
	if err == nil {
		fileLock.Lock()
		_ = os.Remove("data/UI.txt")
		// 创建文件
		file, err := os.OpenFile("data/UI.txt", os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			_, _ = file.WriteString(html)
		}
		fileLock.Unlock()
	}
}
