package api

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"go_crontab/common"
	"go_crontab/service"
	"strings"
)

type CrontabJobApiController struct {
	common.ApiController
}

func(c *CrontabJobApiController) URLMapping(){
	c.Mapping("JobSave",c.JobSave)
	c.Mapping("JobDelete",c.JobDelete)
	c.Mapping("JobList",c.JobList)
	c.Mapping("JobKill",c.JobKill)
}

var(
	err error
	requestBody []byte
	job common.Job
	oldJobData *common.Job
	jobList []*common.Job
)

/**
	新建/编辑任务
 */
// @router /job/jobSave [post]
func (c *CrontabJobApiController) JobSave(){
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"),"application/json"){
		c.ResponseFailed("500","格式不正确")
	}
	requestBody = c.Ctx.Input.RequestBody
	if err = json.Unmarshal(requestBody,&job);err!=nil{
		logs.Error("job save json unmarshal errors:")
		logs.Error(err)
		c.ResponseFailed("500","保存时解析json出错")
	}
	if job.Name == "" || job.CronExpr == "" || job.Command == ""{
		c.ResponseFailed("500","传入的数据不完整")
	}
	if oldJobData , err = service.G_jobMgr.SaveJob(&job);err != nil{
		logs.Error(err)
		c.ResponseFailed("500","保存时出现网络错误，可能是单点故障")
	}
	c.ResponseSuccess(oldJobData,"任务操作成功")
}

/**
	删除任务
 */
// @router /job/jobDelete [post]
func (c *CrontabJobApiController) JobDelete(){
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"),"application/json"){
		c.ResponseFailed("500","格式不正确")
	}
	requestBody = c.Ctx.Input.RequestBody
	if err = json.Unmarshal(requestBody,&job);err!=nil{
		logs.Error("job save json unmarshal errors:")
		logs.Error(err)
		c.ResponseFailed("500","删除时解析json出错")
	}
	if job.Name == ""{
		c.ResponseFailed("500","传入的数据不完整")
	}
	if oldJobData , err = service.G_jobMgr.DeleteJob(job.Name);err != nil{
		logs.Error(err)
		c.ResponseFailed("500","删除时出现网络错误，可能是单点故障")
	}
	c.ResponseSuccess(oldJobData,"删除任务操作成功")
}

/**
	获取任务列表
 */
// @router /job/jobList [get]
func (c *CrontabJobApiController) JobList(){
	if jobList , err = service.G_jobMgr.GetListJob();err !=nil{
		logs.Error(err)
		c.ResponseFailed("500","获取列表时出现网络错误，可能是单点故障")
	}
	c.ResponseSuccess(jobList,"获取任务列表成功")
}

/**
	强杀任务
 */
// @router /job/jobKill [post]
func (c * CrontabJobApiController) JobKill(){
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"),"application/json"){
		c.ResponseFailed("500","格式不正确")
	}
	requestBody = c.Ctx.Input.RequestBody
	if err = json.Unmarshal(requestBody,&job);err!=nil{
		logs.Error("job save json unmarshal errors:")
		logs.Error(err)
		c.ResponseFailed("500","强杀任务时解析json出错")
	}
	if job.Name == ""{
		c.ResponseFailed("500","传入的数据不完整")
	}
	if err = service.G_jobMgr.KillJob(job.Name);err != nil{
		logs.Error(err)
		c.ResponseFailed("500","强杀时出现网络错误，可能是单点故障")
	}
	c.ResponseSuccess(nil,"强杀任务操作成功")
}

