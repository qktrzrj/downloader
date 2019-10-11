package main

import (
	"github.com/andlabs/ui"
	"strconv"
)

type modelHandler struct {
	lab    int
	colNum int
}

func (modelhandler *modelHandler) ColumnTypes(m *ui.TableModel) []ui.TableValue {
	tValue := make([]ui.TableValue, modelhandler.colNum)
	maxCol := modelhandler.colNum
	if modelhandler.lab == 0 {
		maxCol -= 3
		tValue[6] = ui.TableString("")
		tValue[7] = ui.TableString("")
		tValue[8] = ui.TableString("")
	}
	for idx := 0; idx < maxCol; idx++ {
		tValue[idx] = ui.TableString("") // Init strings columns
	}
	return tValue
}

func (modelhandler *modelHandler) NumRows(m *ui.TableModel) int {
	if modelhandler.lab == 0 {
		if len(Downloader.ActiveRowToTaskId) > 0 {
			return len(Downloader.ActiveRowToTaskId) - 1
		}
	}
	if len(Downloader.CompleteRowToTaskId) > 0 {
		return len(Downloader.CompleteRowToTaskId) - 1
	}
	return 0
}

func (modelhandler *modelHandler) CellValue(m *ui.TableModel, row, column int) ui.TableValue {
	maxCol := modelhandler.colNum
	var task *Task
	if modelhandler.lab == 0 {
		if row >= len(Downloader.ActiveRowToTaskId) {
			return ui.TableString("")
		}
		id := Downloader.ActiveRowToTaskId[row]
		task = Downloader.ActiveTaskMap[id]
		maxCol -= 3
		switch column {
		case 4:
			return ui.TableInt(task.downloadCount * 100 / task.fileLength)
		case 5:
			if task.speedCount >= 1024 {
				return ui.TableString(strconv.FormatFloat(task.speedCount/1024, 'f', 2, 64) + "m/s")
			}
			return ui.TableString(strconv.FormatFloat(task.speedCount, 'f', 2, 64) + "k/s")
		case 6:
			if task.speedCount == 0 {
				return ui.TableString("---")
			}
			if task.remainingTime >= 60*60*24 {
				return ui.TableString(strconv.FormatFloat(task.remainingTime/60*60*24, 'f', 2, 64) + " d")
			}
			if task.remainingTime >= 60*60 {
				return ui.TableString(strconv.FormatFloat(task.remainingTime/60*60, 'f', 2, 64) + " h")
			}
			if task.remainingTime >= 60 {
				return ui.TableString(strconv.FormatFloat(task.remainingTime/60, 'f', 2, 64) + " m")
			}
			return ui.TableString(strconv.FormatFloat(task.remainingTime, 'f', 2, 64) + " s")
		case 7:
			return ui.TableString(DpStatusMap[task.Status])
		case 8:
			return ui.TableString("取消")
		}
	} else {
		if row >= len(Downloader.CompleteRowToTaskId) {
			return ui.TableString("")
		}
		id := Downloader.CompleteRowToTaskId[row]
		task = Downloader.CompleteTaskMap[id]
		switch column {
		case 4:
			return ui.TableString("打开")
		case 5:
			return ui.TableString("删除")
		}
	}
	if column < maxCol {
		switch column {
		case 0:
			return ui.TableString(strconv.Itoa(row + 1))
		case 1:
			return ui.TableString(task.FileName)
		case 2:
			return ui.TableString(task.SavePath)
		case 3:
			return ui.TableString(task.Url)
		}
	}
	return nil
}

func (modelhandler *modelHandler) SetCellValue(m *ui.TableModel, row, column int, value ui.TableValue) {
	maxCol := modelhandler.colNum
	var task *Task
	var id taskId
	if modelhandler.lab == 0 {
		if row >= len(Downloader.ActiveRowToTaskId) {
			return
		}
		id = Downloader.ActiveRowToTaskId[row]
		task = Downloader.ActiveTaskMap[id]
	} else {
		if row >= len(Downloader.CompleteRowToTaskId) {
			return
		}
		id = Downloader.CompleteRowToTaskId[row]
		task = Downloader.CompleteTaskMap[id]
	}
	if column < maxCol {
		switch column {
		case 4:
			// 打开
			event := DownloadEvent{
				TaskId: id,
				Enum:   Open,
			}
			Downloader.Event <- event
		case 5:
			// 删除
			event := DownloadEvent{
				TaskId: id,
				Enum:   Remove,
			}
			Downloader.Event <- event
		case 7:
			if task.Status == Downloading || task.Status == Waiting {
				// 暂停
				event := DownloadEvent{
					TaskId: id,
					Enum:   Pause,
				}
				Downloader.Event <- event
			} else {
				// 继续下载
				event := DownloadEvent{
					TaskId: id,
					Enum:   Resume,
				}
				Downloader.Event <- event
			}
		case 8:
			// 取消
			event := DownloadEvent{
				TaskId: id,
				Enum:   Cancel,
			}
			Downloader.Event <- event
		}
	}
}

func newModelHandler(lab, col int) *modelHandler {
	modelh := new(modelHandler)
	modelh.lab = lab
	modelh.colNum = col
	return modelh
}
