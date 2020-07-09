package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"] = append(beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"],
        beego.ControllerComments{
            Method: "CheckCronExpr",
            Router: "/job/checkCronExpr",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"] = append(beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"],
        beego.ControllerComments{
            Method: "JobDelete",
            Router: "/job/jobDelete",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"] = append(beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"],
        beego.ControllerComments{
            Method: "JobKill",
            Router: "/job/jobKill",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"] = append(beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"],
        beego.ControllerComments{
            Method: "JobList",
            Router: "/job/jobList",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"] = append(beego.GlobalControllerRouter["go_crontab/api:CrontabJobApiController"],
        beego.ControllerComments{
            Method: "JobSave",
            Router: "/job/jobSave",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
