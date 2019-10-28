package common

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

var (
	DB         *sql.DB
	TaskInsert *sql.Stmt
	TaskUpdate *sql.Stmt
	TaskDelete *sql.Stmt
	SegInsert  *sql.Stmt
	SegDelete  *sql.Stmt
	UIInsert   *sql.Stmt
	DBLock     sync.Mutex
)

// task status : 0-未完成,1-完成,2-出错
const (
	INCOMPLETE = iota
	SUCCESS
	ERRORED
)

func init() {
	DB, err := sql.Open("sqlite3", "data/downloader.db")
	if err != nil {
		log.Fatalln("数据库启动失败")
	}

	sql_table := `
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
		);
		create table if not exists segment
		(
			task_id varchar(255),
			start int,
			end int,
			finish int
		);
		create table if not exists ui
		(
			content text
		);`

	_, _ = DB.Exec(sql_table)

	TaskInsert, err = DB.Prepare(`INSERT INTO task(id,renewal,status,file_length,final_link,file_name,save_path) values(?,?,?,?,?,?,?)`)
	if err != nil {
		log.Fatalln("任务插入准备失败!")
	}
	TaskUpdate, err = DB.Prepare(`update task set status=? where id=?`)
	if err != nil {
		log.Fatalln("任务更新准备失败!")
	}
	TaskDelete, err = DB.Prepare(`DELETE FROM task WHERE id=?`)
	if err != nil {
		log.Fatalln("任务删除准备失败!")
	}
	SegInsert, err = DB.Prepare(`INSERT into segment(task_id,start,end,finish) values (?,?,?,?)`)
	if err != nil {
		log.Fatalln("片段插入准备失败!")
	}
	SegDelete, err = DB.Prepare(`delete from segment where task_id=?`)
	if err != nil {
		log.Fatalln("片段删除准备失败!")
	}
	UIInsert, err = DB.Prepare(`insert into ui(content) values (?) `)
	if err != nil {
		log.Fatalln("片段删除准备失败!")
	}
}
