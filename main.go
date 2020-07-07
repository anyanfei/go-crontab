package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/joho/godotenv"
	_ "go_crontab/routers"
	"go_crontab/service"
	"os"
	"runtime"
)

func init()  {
	var err error
	//根据cpu核心数配置运行时使用的内核数
	runtime.GOMAXPROCS(runtime.NumCPU())
	//加载文件
	if err = godotenv.Load(".env");err !=nil{
		logs.Error("no .env file")
	}
	//初始化etcd
	if err = service.InitJobMgr();err !=nil{
		logs.Error("初始化etcd报错")
		logs.Error(err)
	}
}

func main() {
	//把请求复制过来，不然获取不到请求的内容
	beego.BConfig.CopyRequestBody = true
	beego.Run(":"+os.Getenv("API_PORT"))
}

