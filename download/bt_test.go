package download

import (
	"database/sql"
	"downloader/common"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestBt__downSeg(t *testing.T) {
	url := "https://www.sqlite.org/2019/sqlite-dll-win64-x64-3300100.zip"
	common.DB, _ = sql.Open("sqlite3", "conf/downloader.db")
	Download = Downloader{
		MaxRoutineNum:    1,
		SegSize:          500 * 1024,
		BufferSize:       0,
		SavePath:         "./",
		MaxActiveTaskNum: 3,
	}
	Download.Init()
	go Download.ListenEvent()
	g := common.NewClient()
	fileInfo, err := common.GetFileInfo(url, g)
	if err != nil {
		panic(err)
	}
	fileInfo.SavePath = Download.SavePath
	fmt.Println(time.Now())
	_, _ = Download.AddTask(fileInfo, common.NewClient())
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

func Test__parse(t *testing.T) {
	var parseLink = `<a [^*]*href="([*]+)"[^>]*>[^<]*[^/]*[^>]a>`
	resp, _ := http.DefaultClient.Get("https://studygolang.com/dl")
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	reg := regexp.MustCompile(parseLink)
	findAll := reg.FindAllSubmatch(content, -1)
	for _, message := range findAll {
		go func() {
			filelink := string(message[1])
			if strings.Index(filelink, "http") == -1 {
				filelink = resp.Request.Host + filelink
				fmt.Println(filelink)
			}
		}()
	}
}
