package common

import (
	"github.com/go-ini/ini"
	"log"
	"os/user"
	"runtime"
	"strconv"
)

var (
	Cfg        *ini.File
	RunMode    string
	AllPath    string
	RoutineNum int
	MaxTaskNum int32
)

var SavePath = map[string]string{
	"windows": `\Downloads\`,
	"darwin":  `/Download/`,
}

func init() {
	var err error
	Cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("Fail to parse 'conf/app.ini': %v", err)
	}

	RunMode = Cfg.Section("").Key("RUN_MODE").MustString("debug")

	sec, err := Cfg.GetSection("setting")

	if err != nil {
		log.Fatalf("Fail to get section 'setting': %v", err)
	}

	AllPath = sec.Key("SAVE_PATH").MustString(defaultSavePath())
	RoutineNum = sec.Key("ROUTINE_NUM").MustInt(20)
	MaxTaskNum = int32(sec.Key("MAX_TASK_NUM").MustInt(3))
}

func SetValue(path string, routineNum int, maxTaskNum int) {
	sec, err := Cfg.GetSection("setting")

	if err != nil {
		log.Fatalf("Fail to get section 'setting': %v", err)
	}
	sec.Key("SAVE_PATH").SetValue(path)
	sec.Key("ROUTINE_NUM").SetValue(strconv.Itoa(routineNum))
	sec.Key("MAX_TASK_NUM").SetValue(strconv.Itoa(maxTaskNum))
	_ = Cfg.SaveTo("data/app.ini")
	AllPath = path
	RoutineNum = routineNum
}

func defaultSavePath() string {
	var path string
	current, err := user.Current()
	if err == nil {
		path = current.HomeDir + SavePath[runtime.GOOS]
		return path
	}
	if "windows" == runtime.GOOS {
		s, err := HomeWindows()
		if err != nil {
			path = s + SavePath["windows"]
			return path
		}
	}
	// Unix-like system, so just assume Unix
	s, err := HomeUnix()
	if err != nil {
		path = s + SavePath["windows"]
		return path
	}
	return ""
}
