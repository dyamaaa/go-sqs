package main

import (
	"go-sqs/internal/queue"
	"go-sqs/internal/server"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// ログディレクトリのパスを設定
	logDir := "./log"
	logPath := filepath.Join(logDir, "app.log")

	// ログディレクトリが存在しない場合は作成
	// os.Statでディレクトリの存在を確認し、存在しない場合はos.Mkdirでディレクトリを作成
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	// ログファイルを開く
	// os.OpenFileを使用してログファイルを開き、エラーがあればログに出力
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// ログの出力先を設定
	log.SetOutput(logFile)

	// データベースのパスを設定
	dbPath := "./tmp/badger"
	// データベースディレクトリが存在しない場合は作成
	// os.Statでディレクトリの存在を確認し、存在しない場合はos.Mkdirでディレクトリを作成
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err := os.Mkdir(dbPath, 0755)
		if err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	// キューマネージャを作成
	// queue.NewManagerを使用してキューマネージャを作成し、エラーがあればログに出力
	q, err := queue.NewManager(dbPath)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	// サーバーを作成
	// server.NewServerを使用してサーバーを作成
	srv := server.NewServer(q)
	go func() {
		// サーバーを起動
		// srv.Runを使用してサーバーを起動し、エラーがあればログに出力
		if err := srv.Run(":8080"); err != nil {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()
	select {} // 無限ループでプログラムを終了させない
}
