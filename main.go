package main

import (
	"System_Log/function/Get_config"
	"System_Log/function/Log_func"
	"fmt"
)

func main() {
	//存储时间设置的代码日志的
	savetime, err := Get_config.Get_config_int("log", "savetime")
	if err != nil {
		fmt.Println("出错了哦:", err)
	}
	fmt.Println("存储天数您设置的是:", savetime)
	//间隔时间
	retime, err := Get_config.Get_config_int("log", "retime")
	if err != nil {
		fmt.Println("出错了哦:", err)
		return
	}
	//获取存储日志的目录位置
	dir, err := Get_config.Get_config("log", "dir")
	if err != nil {
		fmt.Println("出错了哦:", err)
	}

	//清理的函数
	go func() {
		Log_func.CleanUpOldLogs(dir, savetime)
	}()

	go func() {
		//	记录初次系统的静态的数据
		Log_func.CollectAndLogSystemInfo(dir)
	}()

	Log_func.LogPeriodically(dir, retime)
}
