package service

import "github.com/astaxie/beego"

type ApiController struct {
	beego.Controller
}

func (api *ApiController)ResponseSuccess(data interface{},msg string)  {
	response:=make(map[string]interface{})
	response["code"]= "000"
	response["data"]= data
	response["msg"] = msg
	response["success"] = true
	api.Data["json"] = response
	api.ServeJSON()
	api.StopRun()
}

func (api *ApiController)ResponseFailed(code , msg string)  {
	response := make(map[string]interface{})
	response["code"] = code
	response["data"] = nil
	response["msg"] = msg
	response["success"] = false
	api.Data["json"] = response
	api.ServeJSON()
	api.StopRun()
}
