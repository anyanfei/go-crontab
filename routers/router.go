package routers

import (
	"github.com/astaxie/beego"
	"go_crontab/api"
)

func init() {
	beego.Include(&api.CrontabJobApiController{})
}
