package api

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"go_crontab/service"
	"strings"
)

type CrontabJobApiController struct {
	service.ApiController
}

func(c *CrontabJobApiController) URLMapping(){
	c.Mapping("JobSave",c.JobSave)
	c.Mapping("JobDelete",c.JobDelete)
	c.Mapping("JobList",c.JobList)
	c.Mapping("JobKill",c.JobKill)
	c.Mapping("CheckCronExpr",c.CheckCronExpr)
	c.Mapping("JobLogsList",c.JobLogsList)
}

var(
	err         error
	requestBody []byte
	job         service.Job
	oldJobData  *service.Job
	jobList     []*service.Job
	nextTime    []string
	jobListRequest service.JobListRequest
	keyWord 	string
	getTaskLogs   service.GetTaskLogs
	taskLogsData  map[string]interface{}


)

/**
	检查crontab表达式是否正确
 */
// @router /job/checkCronExpr [post]
func (c *CrontabJobApiController) CheckCronExpr(){
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"),"application/json"){
		c.ResponseFailed("500","格式不正确")
	}
	requestBody = c.Ctx.Input.RequestBody
	if err = json.Unmarshal(requestBody,&job);err!=nil{
		logs.Error("job save json unmarshal errors:")
		logs.Error(err)
		c.ResponseFailed("500","保存时解析json出错")
	}
	if job.CronExpr == ""{
		c.ResponseFailed("500","请传入表达式")
	}
	if err , nextTime = service.CheckCrontabExpr(job.CronExpr);err !=nil{
		c.ResponseFailed("500",err.Error())
	}
	c.ResponseSuccess(nextTime,"获取表达式成功")
}

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
// @router /job/jobList [post]
func (c *CrontabJobApiController) JobList(){
	var (
		allCount int
	)
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"),"application/json"){
		c.ResponseFailed("500","格式不正确")
	}
	requestBody = c.Ctx.Input.RequestBody
	if err = json.Unmarshal(requestBody,&jobListRequest);err!=nil{
		logs.Error("get job list unmarshal errors:")
		logs.Error(err)
		c.ResponseFailed("500","获取列表时解析json出错")
	}
	if jobListRequest.Page == 0{
		c.ResponseFailed("500","页码必传")
	}
	if jobListRequest.PageSize == 0{
		c.ResponseFailed("500","每页个数必传")
	}
	keyWord = ""
	if jobListRequest.KeyWord != ""{
		keyWord = jobListRequest.KeyWord
	}
	if jobList , err = service.G_jobMgr.GetListJob(keyWord);err !=nil{
		logs.Error(err)
		c.ResponseFailed("500","获取列表时出现网络错误，可能是单点故障")
	}
	allCount = len(jobList)
	var tempSlice = make([]interface{},0)
	for _,v := range jobList{
		tempSlice = append(tempSlice,*v)
	}
	var resultList,resList []interface{}
	tempSliceChunk := service.SliceChunk(tempSlice,jobListRequest.PageSize)
	indexNum := jobListRequest.Page - 1
	if indexNum < len(tempSliceChunk){
		resultList = tempSliceChunk[indexNum]
		for _,vv := range resultList{
			if vv != nil{
				resList = append(resList,vv)
			}
		}
	}
	var resultData = make(map[string]interface{})
	resultData["all_count"] = allCount
	resultData["lists"] = resList
	c.ResponseSuccess(resultData,"获取任务列表成功")
}

/**
	强杀任务(暂时不用)
 */
// @router /job/jobKill [post]
func (c *CrontabJobApiController) JobKill(){
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

/**
	获取日志信息（分页）
 */
// @router /job/logs [post]
func (c *CrontabJobApiController) JobLogsList(){
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"),"application/json"){
		c.ResponseFailed("500","格式不正确")
	}
	requestBody = c.Ctx.Input.RequestBody
	if err = json.Unmarshal(requestBody,&getTaskLogs);err!=nil{
		logs.Error(err)
		c.ResponseFailed("500","获取日志任务时解析json出错")
	}
	if getTaskLogs.JobName == ""{
		c.ResponseFailed("500","传入的数据不完整")
	}
	taskLogsData = service.GetTaskLogsByPage(getTaskLogs.Page,getTaskLogs.PageSize,getTaskLogs.JobName)

	c.ResponseSuccess(taskLogsData,"获取列表成功")
}

