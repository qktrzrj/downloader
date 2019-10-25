package main

import (
	"database/sql"
	"downloader/common"
	"downloader/downloader"
	"downloader/routers"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	//url := "https://download.jetbrains.8686c.com/idea/ideaIC-2019.2.2.dmg"
	common.DB, _ = sql.Open("sqlite3", "conf/downloader.db")
	downloader.Download = downloader.Downloader{
		MaxRoutineNum:    common.RoutineNum,
		SegSize:          500 * 1024,
		SavePath:         common.AllPath,
		MaxActiveTaskNum: common.MaxTaskNum,
	}
	downloader.Download.Init()
	go downloader.Download.ListenEvent()

	router := routers.InitRouter()

	server := &http.Server{
		Addr:         ":4800",
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
	}
	log.Println("Server exiting")
}
