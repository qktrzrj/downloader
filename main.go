package main

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

var (
	db         *sql.DB
	model      *ui.TableModel
	allPath    string
	routineNum int
	maxTaskNum int32
)

var savePath = map[string]string{
	"windows": "\\Downloads\\",
	"darwin":  "\\Download\\",
	"linux":   "\\Download\\",
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
	routineNum = 20
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
	Downloader = downloader{
		MaxRoutineNum:    routineNum,
		SegSize:          100 * 1024,
		SavePath:         allPath,
		maxActiveTaskNum: maxTaskNum,
	}
	Downloader.Init()
	go Downloader.ListenEvent()
	_ = ui.Main(SetUI())
}
