package main

import (
	"fmt"
	"github.com/andlabs/ui"
)

var (
	rowCount int
	mh       = make(map[int]*modelHandler)
	model    = make(map[int]*ui.TableModel)
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
	mh[0] = newModelHandler(0, 7)
	model[0] = ui.NewTableModel(mh[0])
	table := ui.NewTable(&ui.TableParams{
		Model:                         model[0],
		RowBackgroundColorModelColumn: 2,
	})
	vbox.Append(table, true)
	// 初始化列
	table.AppendTextColumn("", 0, ui.TableModelColumnNeverEditable, &ui.TableTextColumnOptionalParams{ColorModelColumn: -1})
	table.AppendTextColumn("文件名", 1, ui.TableModelColumnAlwaysEditable, nil)
	table.AppendTextColumn("保存路径", 2, ui.TableModelColumnAlwaysEditable, nil)
	table.AppendTextColumn("url", 3, ui.TableModelColumnAlwaysEditable, nil)
	table.AppendProgressBarColumn("进度", 4)
	table.AppendButtonColumn("", 5, ui.TableModelColumnAlwaysEditable)
	table.AppendButtonColumn("", 6, ui.TableModelColumnAlwaysEditable)

	downloadButton.OnClicked(func(button *ui.Button) {
		if input.Text() == "" {
			ui.MsgBoxError(mainwin, "错误", "请输入下载链接!")
			return
		}
		fileId, err := getFileInfo(input.Text())
		if err != nil {
			ui.MsgBoxError(mainwin, "错误", fmt.Sprintln(err))
			return
		}
		input.SetText("")
		fileQueue.Enqueue(fileId)
		sch <- struct{}{}
		mh[0].RowFileId = append(mh[0].RowFileId, fileId)
		rowCount++
		model[0].RowInserted(rowCount - 1)
	})
	return vbox
}
