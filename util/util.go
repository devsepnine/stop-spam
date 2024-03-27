package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Config struct {
	DiscordToken  string   `json:"discord_token"`
	SummonTimeout int      `json:"summon_timeout"`
	WhiteList     []string `json:"white_list"`
	LimitCount    int      `json:"limit_count"`
}

const (
	filePath = "./config/config.json"
)

var config Config

var once sync.Once

func GetConfig() Config {
	once.Do(func() {
		configFile, err := os.Open(filePath)
		defer func(configFile *os.File) {
			err := configFile.Close()
			if err != nil {
				return
			}
		}(configFile)
		if err != nil {
			panic(err)
		}
		j := json.NewDecoder(configFile)
		err = j.Decode(&config)
	})
	return config
}

func UpdateConfig(key string, value interface{}) error {
	fmt.Printf("type %v \n", value)

	switch value.(type) {
	//case string:
	//	//tv := value.(string)
	//	//config.WhiteList = tv
	case int:
		tv := value.(int)
		config.SummonTimeout = tv
	}

	configFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("파일을 읽을 수 없습니다. : %v", err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(configFile, &data)
	if err != nil {
		return fmt.Errorf("JSON 데이터를 읽을 수 없습니다 : %v", err)
	}

	data[key] = value

	updatedData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("fail Marshal %v", err)
	}

	err = os.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("파일 업데이트에 실패 하였습니다. %v", err)
	}

	return nil
}
