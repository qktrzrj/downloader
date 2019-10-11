package main

import "github.com/andlabs/ui"

var (
	cpMh    *modelHandler
	CpModel *ui.TableModel
)

func completePage() ui.Control {
	// 构建完成界面主窗口
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	// 搜索框
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)
	// 下载输入框和按钮
	input := ui.NewSearchEntry()
	downloadButton := ui.NewButton("搜 索")
	hbox.Append(input, true)
	hbox.Append(downloadButton, false)
	cpMh = newModelHandler(1, 6)
	CpModel = ui.NewTableModel(cpMh)
	table := ui.NewTable(&ui.TableParams{
		Model:                         CpModel,
		RowBackgroundColorModelColumn: -1,
	})
	vbox.Append(table, true)
	// 初始化列
	table.AppendTextColumn("", 0, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("文件名", 1, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("保存路径", 2, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("url", 3, ui.TableModelColumnNeverEditable, nil)
	table.AppendButtonColumn("", 4, ui.TableModelColumnAlwaysEditable)
	table.AppendButtonColumn("", 5, ui.TableModelColumnAlwaysEditable)
	return vbox
}
