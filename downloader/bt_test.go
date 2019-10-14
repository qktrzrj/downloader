package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBt__downSeg(t *testing.T) {
	url := "https://www.typora.io/windows/typora-setup-x64.exe"
	Download = Downloader{
		MaxRoutineNum:    200,
		SegSize:          100 * 1024,
		BufferSize:       0,
		SavePath:         "F:/goProject/Downloader/",
		MaxActiveTaskNum: 3,
	}
	Download.Init()
	g := newGout()
	fileInfo, err := GetFileInfo(url, g)
	if err != nil {
		panic(err)
	}
	_ = Download.AddTask(fileInfo, g)
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
