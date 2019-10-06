package ui

import (
	"downloader/downloader"
	"fmt"
	"github.com/andlabs/ui"
)

var (
	rowCount int
	dpMh     *modelHandler
	DpModel  *ui.TableModel
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
	downloadButton := ui.NewButton("下载")
	hbox.Append(input, true)
	hbox.Append(downloadButton, false)
	dpMh = newModelHandler(0, 7)
	DpModel = ui.NewTableModel(dpMh)
	table := ui.NewTable(&ui.TableParams{
		Model:                         DpModel,
		RowBackgroundColorModelColumn: 2,
	})
	vbox.Append(table, true)
	// 初始化列
	table.AppendTextColumn("", 0, ui.TableModelColumnNeverEditable,
		&ui.TableTextColumnOptionalParams{ColorModelColumn: -1})
	table.AppendTextColumn("文件名", 1, ui.TableModelColumnAlwaysEditable, nil)
	table.AppendTextColumn("保存路径", 2, ui.TableModelColumnAlwaysEditable, nil)
	table.AppendTextColumn("url", 3, ui.TableModelColumnAlwaysEditable, nil)
	table.AppendProgressBarColumn("进度", 4)
	table.AppendButtonColumn("", 5, ui.TableModelColumnAlwaysEditable)
	table.AppendButtonColumn("", 6, ui.TableModelColumnAlwaysEditable)

	downloadButton.OnClicked(func(button *ui.Button) {
		url := input.Text()
		if url == "" {
			ui.MsgBoxError(MainWin, "错误", "请输入下载链接!")
			return
		}
		task, err := downloader.GetFileTask(url, downloader.NewHostClient(url))
		if err != nil {
			ui.MsgBoxError(MainWin, "错误", fmt.Sprintln(err))
			return
		}
		// 当文件名重复时
		if !downloader.FileIsNotExist(task.SavePath + "/" + task.FileName) {
			window := ui.NewWindow("提示", 400, 200, false)
			window.SetMargined(true)
			box := ui.NewVerticalBox()
			window.SetChild(box)
			box.Append(ui.NewLabel("文件名已存在,请重命名！"), false)
			nbox := ui.NewHorizontalBox()
			vbox.Append(nbox, false)
			name := ui.NewEntry()
			getName := ui.NewButton("保存")
			nbox.Append(name, true)
			nbox.Append(getName, false)
			getName.OnClicked(func(innerButton *ui.Button) {
				if !downloader.FileIsNotExist(task.SavePath + "/" + name.Text()) {
					ui.MsgBox(window, "错误", "文件名重复！")
				}
				task.FileName = name.Text()
				window.Destroy()
			})
			window.Show()
		}
		err = downloader.Downloader.AddTask(task)
		if err != nil {
			ui.MsgBoxError(MainWin, "错误", fmt.Sprintln(err))
			return
		}
		main.downloadDirect(fileId)
		input.SetText("")
		main.fileQueue.Enqueue(fileId)
		dpmh[0].RowFileId = append(dpmh[0].RowFileId, fileId)
		rowCount++
		model[0].RowInserted(rowCount - 1)
		main.sch <- struct{}{}
	})
	return vbox
}
