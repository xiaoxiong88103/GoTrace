package Get_system

import (
	"System_Log/function/Get_config"
	"fmt"
	"strconv"
	"strings"
)

// PrintFilteredValues 根据 Filterate_proc 的结果打印相应的值
func PrintFilteredValues(pidname string, outputBuilder *strings.Builder) error {
	// 获取筛选后的关键字
	filteredKeys, err := Get_config.Filterate_proc()
	if err != nil {
		return err
	}

	// 获取 proc_name 信息
	cpu, mem, runtime, fd, version, pid, err := Get_proc_info(pidname)
	if err != nil {
		return err
	}

	// 创建一个键值对映射
	infoMap := map[string]interface{}{
		"cpu":     cpu,
		"mem":     mem,
		"runtime": runtime,
		"fd":      fd,
		"version": version,
		"pid":     pid,
	}

	// 预定义关键字的顺序
	predefinedOrder := []string{"cpu", "mem", "runtime", "fd", "version", "pid"}

	// 为每个键设置不同的固定长度
	keyLengths := map[string]int{
		"cpu":     6,
		"mem":     9,
		"runtime": 14,
		"fd":      6,
		"version": 12,
		"pid":     8,
	}

	// 遍历预定义的顺序，检查filteredKeys是否包含该关键字，如果是，则累积到输出字符串
	for _, key := range predefinedOrder {
		for _, filteredKey := range filteredKeys {
			cleanKey := strings.TrimSuffix(filteredKey, ":")
			if cleanKey == key {
				if value, exists := infoMap[cleanKey]; exists {
					length, ok := keyLengths[cleanKey]
					if !ok {
						length = 10 // 默认长度，如果没有特别指定
					}
					formattedValue := fmt.Sprintf("%-*v", length, value) // 使用指定长度对齐
					if outputBuilder.Len() > 0 {
						outputBuilder.WriteString(" | ")
					}
					outputBuilder.WriteString(fmt.Sprintf("%s: %s", cleanKey, formattedValue))
					break // 找到匹配项后跳出内层循环
				}
			}
		}
	}

	return nil

	// 遍历预定义的顺序，检查filteredKeys是否包含该关键字，如果是，则累积到输出字符串
	//for _, key := range predefinedOrder {
	//	for _, filteredKey := range filteredKeys {
	//		cleanKey := strings.TrimSuffix(filteredKey, ":")
	//		if cleanKey == key {
	//			if value, exists := infoMap[cleanKey]; exists {
	//				if outputBuilder.Len() > 0 {
	//					outputBuilder.WriteString(" | ") // 在现有内容后添加分隔符
	//				}
	//				outputBuilder.WriteString(fmt.Sprintf("%s: %v", cleanKey, value))
	//				break // 找到匹配项后跳出内层循环
	//			}
	//		}
	//	}
	//}

}

// PrintSystemInfo 打印系统信息，基于 Filterate_system 的结果
func PrintSystemInfo(outputBuilder *strings.Builder) error {
	// 获取经过筛选的系统关键字
	filteredKeys, err := Get_config.Filterate_system()
	if err != nil {
		return err
	}

	// 获取系统信息
	cpu, free, loadavg, err := System_get()
	if err != nil {
		return err
	}

	// 获取时间信息
	uptime, nowtime, err := Get_time()
	if err != nil {
		return err
	}
	// 使用 fmt.Sprintf 格式化输出，保留小数点前两位，并将结果转换为字符串
	formattedCPU := fmt.Sprintf("%.2f", cpu)
	cpuFloat, err := strconv.ParseFloat(formattedCPU, 64)
	if err != nil {
		return err
	}

	//获取磁盘剩余的空间
	fdisk, err := GetDiskUsageInfo()
	if err != nil {
		return err
	}

	// 创建一个键值对映射
	infoMap := map[string]interface{}{
		"sys_cpu": cpuFloat,
		"free":    free,
		"loadavg": loadavg,
		"uptime":  uptime.Format("2006-01-02 15:04:05"),
		"nowtime": nowtime.Format("2006-01-02 15:04:05"),
		"npu":     GetNPULoad(),
		"gpu":     GetGPULoad(),
		"disk":    fdisk,
		"io":      GetIOStats(),
		"network": GetNetworkStats(),
		"rga":     ReadAndFormatRGALoad(),
	}

	// 预定义关键字的顺序
	predefinedOrder := []string{"sys_cpu", "free", "loadavg", "uptime", "nowtime", "io", "network", "npu", "rga", "gpu", "disk"}

	// 为每个键设置不同的固定长度
	keyLengths := map[string]int{
		"sys_cpu": 6,
		"free":    6,
		"loadavg": 6,
		"uptime":  20,
		"nowtime": 20,
		"io":      10,
		"network": 8,
		"npu":     8,
		"rga":     28,
		"gpu":     10,
		"disk":    15,
	}

	// 遍历预定义的顺序，检查filteredKeys是否包含该关键字，如果是，则累积到输出字符串
	for _, key := range predefinedOrder {
		for _, filteredKey := range filteredKeys {
			cleanKey := strings.TrimSuffix(filteredKey, ":")
			if cleanKey == key {
				if value, exists := infoMap[cleanKey]; exists {
					length, ok := keyLengths[cleanKey]
					if !ok {
						length = 10 // 默认长度，如果没有特别指定
					}
					formattedValue := fmt.Sprintf("%-*v", length, value) // 使用指定长度对齐
					if outputBuilder.Len() > 0 {
						outputBuilder.WriteString(" | ")
					}
					outputBuilder.WriteString(fmt.Sprintf("%s: %s", cleanKey, formattedValue))
					break // 找到匹配项后跳出内层循环
				}
			}
		}
	}

	return nil
	//// 遍历预定义的顺序，检查filteredKeys是否包含该关键字，如果是，则累积到输出字符串
	//for _, key := range predefinedOrder {
	//	for _, filteredKey := range filteredKeys {
	//		cleanKey := strings.TrimSuffix(filteredKey, ":")
	//		if cleanKey == key {
	//			if value, exists := infoMap[cleanKey]; exists {
	//				if outputBuilder.Len() > 0 {
	//					outputBuilder.WriteString(" | ") // 在现有内容后添加分隔符
	//				}
	//				outputBuilder.WriteString(fmt.Sprintf("%s: %v", cleanKey, value))
	//				break // 找到匹配项后跳出内层循环
	//			}
	//		}
	//	}
	//}
}
