package main

import (
	"github.com/andlabs/ui"
	"os"
	"path/filepath"
	"strings"
)

var (
	slider *ui.Slider
	path   *ui.Entry
)

func settingPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	lable1 := ui.NewLabel("设置单任务并发数:")
	vbox.Append(lable1, false)
	slider = ui.NewSlider(0, 100)
	slider.SetValue(routineNum)
	vbox.Append(slider, false)
	lable2 := ui.NewLabel("设置保存路径")
	vbox.Append(lable2, false)
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)
	path = ui.NewEntry()
	path.SetText(allPath)
	hbox.Append(path, true)
	open := ui.NewButton("打开")
	hbox.Append(open, false)
	lable3 := ui.NewLabel("设置最大下载数")
	vbox.Append(lable3, false)
	spinbox := ui.NewSpinbox(1, 10)
	spinbox.SetValue(int(maxTaskNum))
	vbox.Append(spinbox, false)

	slider.OnChanged(func(i *ui.Slider) {
		Downloader.MaxRoutineNum = slider.Value()
	})

	open.OnClicked(func(button *ui.Button) {
		s := path.Text()
		if s == "" {
			fileInfo, _ := os.Stat("./")
			s, _ = filepath.Abs(filepath.Dir(fileInfo.Name()))
		}
		s = ui.OpenFile(MainWin)
		if s != "" {
			path.SetText(s[:strings.LastIndex(s, "\\")+1])
		}
	})

	spinbox.OnChanged(func(spinbox *ui.Spinbox) {
		Downloader.maxActiveTaskNum = int32(spinbox.Value())
	})
	return vbox
}
