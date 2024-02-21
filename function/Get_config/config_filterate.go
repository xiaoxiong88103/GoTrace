package Get_config

import (
	"strings"
)

func Filterate_proc() ([]string, error) {
	prockey := []string{
		"version:",
		"pid:",
		"fd:",
		"cpu:",
		"mem:",
		"runtime:",
	}

	section := "pid"

	// 创建一个映射来存储需要检查的keys
	keysToCheck := make(map[string]bool)
	for _, key := range prockey {
		cleanedKey := strings.TrimSuffix(key, ":")
		val, err := Get_config(section, cleanedKey)
		if err != nil {
			return nil, err
		}
		keysToCheck[cleanedKey] = (val != "0")
	}

	// 根据keysToCheck映射重建prockey数组
	var updateprockey []string
	for key, include := range keysToCheck {
		if include {
			updateprockey = append(updateprockey, key+":")
		}
	}

	return updateprockey, nil
}

func Filterate_system() ([]string, error) {
	systemkey := [...]string{
		"sys_cpu:",
		"free:",
		"loadavg:",
		"uptime:",
		"nowtime:",
		"npu:",
		"gpu:",
		"disk:",
	}

	section := "system"

	// 创建一个映射来存储需要检查的keys
	keysToCheck := make(map[string]bool)
	for _, key := range systemkey {
		cleanedKey := strings.TrimSuffix(key, ":")
		val, err := Get_config(section, cleanedKey)
		if err != nil {
			return nil, err
		}
		keysToCheck[cleanedKey] = (val != "0")
	}

	// 根据keysToCheck映射重建system数组
	var updatedSystemkey []string
	for key, include := range keysToCheck {
		if include {
			updatedSystemkey = append(updatedSystemkey, key+":")
		}
	}

	return updatedSystemkey, nil
}
