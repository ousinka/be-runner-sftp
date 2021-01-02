package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

//初始化
func init() {
	file := "./sftp-" + time.Now().Format("20060102") + ".log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return
}

//加载配置文件
func LoadConf(fileName string) *Conf {
	cfg := &Conf{}
	data, err := ioutil.ReadFile(fileName)
	if err != nil{
		panic(err)
	}

	if err := json.Unmarshal([]byte(data), cfg); err != nil {
		panic(err)
	}
	return cfg
}