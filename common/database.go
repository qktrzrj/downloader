package common

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

var (
	TaskDB     *sql.DB
	SegDB      *sql.DB
	UiDB       *sql.DB
	TaskLock   sync.Mutex
	SegLock    sync.Mutex
	UiLock     sync.Mutex
	TaskInsert *sql.Stmt
	TaskUpdate *sql.Stmt
	TaskDelete *sql.Stmt
	SegInsert  *sql.Stmt
	SegDelete  *sql.Stmt
	UIInsert   *sql.Stmt
	UIUpdate   *sql.Stmt
)

// task status : 0-未完成,1-完成,2-出错
const (
	INCOMPLETE = iota
	SUCCESS
	ERRORED
)

func init() {
	TaskDB, _ = sql.Open("sqlite3", "data/task.db")
	SegDB, _ = sql.Open("sqlite3", "data/seg.db")
	UiDB, _ = sql.Open("sqlite3", "data/ui.db")

	task_sql_table := `
		create table if not exists task
		(
			id VARCHAR(36)
				constraint task_pk
					primary key,
			renewal BOOLEAN,
			status int ,
			file_length int,
			final_link varchar(255),
			file_name varchar(255),
			save_path varchar(255)
		);`
	seg_sql_table := `
		create table if not exists segment
		(
			task_id varchar(255),
			start int,
			end int,
			finish int
		);`
	ui_sql_table := `
		create table if not exists ui
		(
			content text
		);`

	_, _ = TaskDB.Exec(task_sql_table)
	_, _ = SegDB.Exec(seg_sql_table)
	_, _ = UiDB.Exec(ui_sql_table)

	var err error
	TaskInsert, err = TaskDB.Prepare(`INSERT INTO task(id,renewal,status,file_length,final_link,file_name,save_path) values(?,?,?,?,?,?,?)`)
	if err != nil {
		log.Fatalln("任务插入准备失败!")
	}
	TaskUpdate, err = TaskDB.Prepare(`update task set status=? where id=?`)
	if err != nil {
		log.Fatalln("任务更新准备失败!")
	}
	TaskDelete, err = TaskDB.Prepare(`DELETE FROM task WHERE id=?`)
	if err != nil {
		log.Fatalln("任务删除准备失败!")
	}
	SegInsert, err = SegDB.Prepare(`INSERT into segment(task_id,start,end,finish) values (?,?,?,?)`)
	if err != nil {
		log.Fatalln("片段插入准备失败!")
	}
	SegDelete, err = SegDB.Prepare(`delete from segment where task_id=?`)
	if err != nil {
		log.Fatalln("片段删除准备失败!")
	}
	UIInsert, err = UiDB.Prepare(`insert into ui(content) values (?) `)
	if err != nil {
		log.Fatalln("UI插入准备失败!")
	}
	UIUpdate, err = UiDB.Prepare(`update ui  set content=?`)
	if err != nil {
		log.Fatalln("UI插入准备失败!")
	}
	rows, err := UiDB.Query("select * from ui")
	if err != nil {
		log.Fatalln("UI插入准备失败!")
	}
	if !rows.Next() {
		insertUI("")
	}
}

func InsertTask(id string, renewal bool, status int, fileLength int64, finalLink string, fileName string, savePath string) {
	TaskLock.Lock()
	_, _ = TaskInsert.Exec(id, renewal, status, fileLength, finalLink, fileName, savePath)
	TaskLock.Unlock()
}

func UpdateTask(status int, id string) {
	TaskLock.Lock()
	_, _ = TaskUpdate.Exec(status, id)
	TaskLock.Unlock()
}

func DeleteTask(id string) {
	TaskLock.Lock()
	_, _ = TaskDelete.Exec(id)
	TaskLock.Unlock()
}

func InsertSeg(id string, start int64, end int64, finish int64) {
	SegLock.Lock()
	_, _ = SegInsert.Exec(id, start, end, finish)
	SegLock.Unlock()
}

func DeleteSeg(id string) {
	SegLock.Lock()
	_, _ = SegDelete.Exec(id)
	SegLock.Unlock()
}

func insertUI(content string) {
	UiLock.Lock()
	_, _ = UIInsert.Exec(content)
	UiLock.Unlock()
}

func UpdateUI(content string) {
	UiLock.Lock()
	_, _ = UIUpdate.Exec(content)
	UiLock.Unlock()
}
