package downloader

import (
	"database/sql"
	"downloader/conf"
	"downloader/util"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestBt__downSeg(t *testing.T) {
	url := "https://www.sqlite.org/2019/sqlite-dll-win64-x64-3300100.zip"
	conf.DB, _ = sql.Open("sqlite3", "conf/downloader.db")
	Download = Downloader{
		MaxRoutineNum:    1,
		SegSize:          500 * 1024,
		BufferSize:       0,
		SavePath:         "./",
		MaxActiveTaskNum: 3,
	}
	Download.Init()
	go Download.ListenEvent()
	g := util.NewClient()
	fileInfo, err := util.GetFileInfo(url, g)
	if err != nil {
		panic(err)
	}
	fileInfo.SavePath = Download.SavePath
	fmt.Println(time.Now())
	_, _ = Download.AddTask(fileInfo, util.NewClient())
	select {}
}

func Test__filePath(t *testing.T) {
	current, _ := user.Current()
	fmt.Println(current.HomeDir)
	fileInfo, _ := os.Stat(current.HomeDir)
	s, _ := filepath.Abs(filepath.Dir(fileInfo.Name()))
	fmt.Println(s)
	run, _ := commands[runtime.GOOS]
	cmd := exec.Command("cmd", "/C", run, s)
	fmt.Println(cmd.String())
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
}
