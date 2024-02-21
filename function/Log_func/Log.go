package Log_func

import (
	"System_Log/function/Get_config"
	"System_Log/function/Get_system"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanUpOldLogs 定期检查并删除旧日志文件，如果没有过期日志则休息1小时
func CleanUpOldLogs(logDir string, maxAgeInDays int) {
	for {
		foundExpiredLogs := false // 标志，用于跟踪是否找到过期日志

		//Log_func.Println("Checking for old Log_func files to clean up...")
		files, err := ioutil.ReadDir(logDir)
		if err != nil {
			//Log_func.Printf("Failed to list files in %s: %v", logDir, err)
			time.Sleep(1 * time.Hour) // 如果无法读取目录，等待1小时再重试
			continue
		}

		now := time.Now()
		for _, file := range files {
			// 解析文件名中的日期
			fileName := file.Name()
			datePart := strings.TrimSuffix(fileName, "_data.log")
			fileDate, err := time.Parse("2006-01-02", datePart)
			if err != nil {
				// 文件名不符合预期格式，跳过
				continue
			}

			// 计算文件日期与当前日期的差异
			if now.Sub(fileDate).Hours() > float64(maxAgeInDays*24) {
				// 如果文件旧于maxAgeInDays天，则删除文件
				if err := os.Remove(filepath.Join(logDir, fileName)); err != nil {
					log.Printf("Failed to delete old Log_func file %s: %v", fileName, err)
				} else {
					log.Printf("Deleted old Log_func file %s", fileName)
					foundExpiredLogs = true // 标记找到并删除了过期日志
				}
			}
		}

		// 如果这次检查没有找到过期的日志文件，休息1小时
		if !foundExpiredLogs {
			log.Println("No expired logs found. Sleeping for 1 hour.")
			time.Sleep(1 * time.Hour)
		} else {
			// 如果找到过期文件，立即进行下一次检查
			time.Sleep(5 * time.Minute) // 按原计划每5分钟检查一次
		}
	}
}

// LogPeriodically 每隔一定时间记录日志到文件
func LogPeriodically(dir string, intervalSeconds int) {
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取当前日期并构建日志文件名
		currentDate := time.Now().Format("2006-01-02")
		logFileName := dir + currentDate + "_data.log"
		// 打开或创建日志文件
		logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("无法打开日志文件:", err)
			continue
		}

		// 初始化一个空字符串来累积输出
		var outputBuilder strings.Builder
		pidname, err := Get_config.Get_config("log", "pid")
		if err != nil {
			fmt.Println("错误了获取pid:", err)
		}
		if pidname == "" {
			// 将累积的字符串写入日志文件
			logFile.WriteString(outputBuilder.String() + "\n")
			logFile.Close()

		} else {
			// 对 proc 信息进行操作 记录到日志
			err = Get_system.PrintFilteredValues(pidname, &outputBuilder)
			if err != nil {
				log.Println("Error processing AI information:", err)
			}

			// 对系统信息进行操作
			err = Get_system.PrintSystemInfo(&outputBuilder)
			if err != nil {
				log.Println("Error processing system information:", err)
			}

			// 将累积的字符串写入日志文件
			logFile.WriteString(outputBuilder.String() + "\n")
			logFile.Close()
		}

	}
}
