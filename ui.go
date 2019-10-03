package main

import (
	"fmt"
	"github.com/andlabs/ui"
)

var mainwin *ui.Window

func addTask(url string) (ui.Control, int, *ui.ProgressBar) {
	// 获取下载文件信息
	id, err := getFileInfo(url)
	if err != nil {
		ui.MsgBoxError(mainwin, "错误", fmt.Sprintln(err))
		return nil, 0, nil
	}
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	hbox.Append(vbox, true)
	fileName := ui.NewLabel(fileMap[id].FileName)
	urlInfo := ui.NewLabel(url)
	bar := ui.NewProgressBar()
	vbox.Append(fileName, false)
	vbox.Append(urlInfo, false)
	vbox.Append(bar, false)
	control := ui.NewButton("暂停")
	delete := ui.NewButton("删除")
	hbox.Append(control, false)
	hbox.Append(delete, false)
	return hbox, id, bar
}

func downloadingPage() ui.Control {
	// 构建下载界面主窗口
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	// 下载框
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)
	// 下载输入框和按钮
	input := ui.NewEntry()
	downloadButton := ui.NewButton("下载")
	hbox.Append(input, true)
	hbox.Append(downloadButton, false)
	// 下载列表
	vbox.Append(ui.NewVerticalSeparator(), false)
	downloadButton.OnClicked(func(button *ui.Button) {
		if input.Text() == "" {
			ui.MsgBoxError(mainwin, "错误", "请输入下载链接!")
			return
		}
		box, fileId, bar := addTask(input.Text())
		if box != nil {
			barList[fileId] = bar
			fileQueue.Enqueue(fileId)
			sch <- struct{}{}
			vbox.Append(box, true)
		}
	})
	return vbox
}

func completePage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	return vbox
}

func settingPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	return vbox
}

func setUI() {
	mainwin = ui.NewWindow("下载器", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})
	tab := ui.NewTab()
	mainwin.SetChild(tab)
	tab.Append("正在下载", downloadingPage())
	tab.SetMargined(0, true)
	tab.Append("下载完成", completePage())
	tab.SetMargined(1, true)
	tab.Append("设置", settingPage())
	tab.SetMargined(2, true)
	mainwin.Show()
}

//func main() {
//	//url := "https://download.jetbrains.8686c.com/idea/ideaIC-2019.2.2.dmg"
//	// 下载器调度启动
//	fileQueue.New()
//	go scheduler()
//	ui.Main(setUI)
//}
