package log

import (
	"log"
	"os"
)

func Log(name string) *log.Logger {
	// 打开文件（不存在就创建），追加模式
	str := "log/" + name + ".log"
	file, err := os.OpenFile(str, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}

	// 设置日志输出到文件
	logger := log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}
