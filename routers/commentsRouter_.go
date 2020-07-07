package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"] = append(beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"],
        beego.ControllerComments{
            Method: "JobSave",
            Router: "/api/jobSave",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
