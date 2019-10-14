package main

import (
	"bytes"
	"database/sql"
	"downloader/downloader"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/therecipe/qt/widgets"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

var (
	db         *sql.DB
	allPath    string
	routineNum int
	maxTaskNum int32
)

var savePath = map[string]string{
	"windows": `\Downloads\`,
	"darwin":  `\Download\`,
	"linux":   `\Download\`,
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
		SegSize:          100 * 1024,
		SavePath:         allPath,
		MaxActiveTaskNum: maxTaskNum,
	}
	downloader.Download.Init()
	go downloader.Download.ListenEvent()
	// needs to be called once before you can start using the QWidgets
	app := widgets.NewQApplication(len(os.Args), os.Args)

	// create a window
	// with a minimum size of 250*200
	// and sets the title to "Hello Widgets Example"
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(250, 200)
	window.SetWindowTitle("Hello Widgets Example")

	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(widget)

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	input := widgets.NewQLineEdit(nil)
	input.SetPlaceholderText("Write something ...")
	widget.Layout().AddWidget(input)

	// create a button
	// connect the clicked signal
	// and add it to the central widgets layout
	button := widgets.NewQPushButton2("and click me!", nil)
	button.ConnectClicked(func(bool) {
		widgets.QMessageBox_Information(nil, "OK", input.Text(), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(button)

	// make the window visible
	window.Show()

	// start the main Qt event loop
	// and block until app.Exit() is called
	// or the window is closed by the user
	app.Exec()
}
