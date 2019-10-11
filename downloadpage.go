package main

import (
	"fmt"
	"github.com/andlabs/ui"
	"strings"
)

var (
	dpMh    *modelHandler
	DpModel *ui.TableModel
)

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
	downloadButton := ui.NewButton("下 载")
	hbox.Append(input, true)
	hbox.Append(downloadButton, false)
	dpMh = newModelHandler(0, 9)
	DpModel = ui.NewTableModel(dpMh)
	table := ui.NewTable(&ui.TableParams{
		Model:                         DpModel,
		RowBackgroundColorModelColumn: -1,
	})
	vbox.Append(table, true)
	// 初始化列
	table.AppendTextColumn("", 0, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("文件名", 1, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("保存路径", 2, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("url", 3, ui.TableModelColumnNeverEditable, nil)
	table.AppendProgressBarColumn("进度", 4)
	table.AppendTextColumn("下载速度", 5, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("剩余时间", 6, ui.TableModelColumnNeverEditable, nil)
	table.AppendButtonColumn("", 7, ui.TableModelColumnAlwaysEditable)
	table.AppendButtonColumn("", 8, ui.TableModelColumnAlwaysEditable)

	downloadButton.OnClicked(func(button *ui.Button) {
		url := input.Text()
		input.SetText("")
		if url == "" {
			ui.MsgBoxError(MainWin, "错误", "请输入下载链接!")
			return
		}
		task, err := GetFileTask(url, NewHostClient(url))
		if err != nil {
			ui.MsgBoxError(MainWin, "错误", fmt.Sprintln(err))
			return
		}
		filePath := ui.SaveFile(MainWin)
		if filePath == "" {
			return
		}
		task.SavePath = filePath
		task.FileName = filePath[strings.LastIndex(filePath, "\\")+1:]
		err = Downloader.AddTask(task)
		if err != nil {
			ui.MsgBoxError(MainWin, "错误", fmt.Sprintln(err))
			return
		}
		// 当文件名重复时
		//if Downloader.FileExist(Downloader.SavePath + task.FileName) {
		//	window := ui.NewWindow("提示", 300, 100, false)
		//	window.SetMargined(true)
		//	box := ui.NewVerticalBox()
		//	window.SetChild(box)
		//	box.Append(ui.NewLabel("文件名已存在,请重命名！"), false)
		//	hbox := ui.NewHorizontalBox()
		//	box.Append(hbox, false)
		//	name := ui.NewEntry()
		//	name.SetText(task.FileName)
		//	getName := ui.NewButton("保 存")
		//	hbox.Append(name, true)
		//	hbox.Append(getName, false)
		//	getName.OnClicked(func(innerButton *ui.Button) {
		//		if name.Text() == "" {
		//			ui.MsgBoxError(window, "错误", "文件名错误！")
		//			return
		//		} else if Downloader.FileExist(task.SavePath + name.Text()) {
		//			ui.MsgBoxError(window, "错误", "文件名重复！")
		//			return
		//		} else {
		//			task.FileName = name.Text()
		//			task.SavePath = Downloader.SavePath + task.FileName
		//			err = Downloader.AddTask(task)
		//			if err != nil {
		//				ui.MsgBoxError(MainWin, "错误", fmt.Sprintln(err))
		//				return
		//			}
		//			MainWin.Enable()
		//			window.Destroy()
		//		}
		//	})
		//	window.OnClosing(func(window *ui.Window) bool {
		//		MainWin.Enable()
		//		return true
		//	})
		//	MainWin.Disable()
		//	window.Show()
		//} else {
		//	task.SavePath = Downloader.SavePath + task.FileName
		//	err = Downloader.AddTask(task)
		//	if err != nil {
		//		ui.MsgBoxError(MainWin, "错误", fmt.Sprintln(err))
		//		return
		//	}
		//}
	})
	return vbox
}
