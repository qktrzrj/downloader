package main

import (
	"bytes"
	"database/sql"
	"downloader/downloader"
	"downloader/util"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	db         *sql.DB
	allPath    string
	routineNum int
	maxTaskNum int32
)

var savePath = map[string]string{
	"windows": `/Downloads/`,
	"darwin":  `/Download/`,
	"linux":   `/Download/`,
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func init() {
	routineNum = 40
	maxTaskNum = 3

	current, err := user.Current()
	if err == nil {
		allPath = current.HomeDir + savePath[runtime.GOOS]
		return
	}
	if "windows" == runtime.GOOS {
		s, err := homeWindows()
		if err != nil {
			allPath = s + savePath["windows"]
			return
		}
	}
	// Unix-like system, so just assume Unix
	s, err := homeUnix()
	if err != nil {
		allPath = s + savePath["windows"]
		return
	}
}

func main() {
	//url := "https://download.jetbrains.8686c.com/idea/ideaIC-2019.2.2.dmg"
	//db, _ = sql.Open("sqlite3", "./downloader.db")
	downloader.Download = downloader.Downloader{
		MaxRoutineNum:    routineNum,
		SegSize:          1024 * 1024,
		SavePath:         allPath,
		MaxActiveTaskNum: maxTaskNum,
	}
	downloader.Download.Init()
	go downloader.Download.ListenEvent()

	router := gin.Default()
	router.GET("/getFileInfo", func(context *gin.Context) {
		result := util.NewResult()
		defer context.JSON(http.StatusOK, result)
		fileInfo, err := util.GetFileInfo(context.Query("url"), util.NewClient())
		if err != nil {
			result.Code = -1
			result.Msg = fmt.Sprint(err)
			return
		}
		fileInfo.SavePath = downloader.Download.SavePath
		result.Data = fileInfo
	})

	router.POST("/addTask", func(context *gin.Context) {
		result := util.NewResult()
		defer context.JSON(http.StatusOK, result)
		var fileInfo util.FileInfo
		err := context.BindJSON(&fileInfo)
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

	router.GET("/getTaskInfo", func(c *gin.Context) {
		// change the reqest to websocket model
		conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			http.NotFound(c.Writer, c.Request)
			return
		}
		for {
			//读取ws中的数据
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			i, err := strconv.Atoi(string(message))
			if err != nil {
				break
			}
			id := downloader.TaskId(i)
			task, ok := downloader.Download.ActiveTaskMap[id]
			if !ok {
				break
			}
			marshal, err := json.Marshal(task)
			if err != nil {
				break
			}
			//写入ws数据
			err = conn.WriteMessage(mt, marshal)
			if err != nil {
				break
			}
		}
	})

	server := &http.Server{
		Addr:         ":4800",
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
	}
	log.Println("Server exiting")
}
