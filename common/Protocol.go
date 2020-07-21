package common

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/gorhill/cronexpr"
	"github.com/idoubi/goz"
	"os"
	"strings"
	"time"
)
//保存事件
const JOB_EVENT_SAVE int = 1
//删除事件
const JOB_EVENT_DELETE int = 2

type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cron_expr"`
}

type Response struct {
	Errno string `json:"errno"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

type JobEvent struct {
	EventType int
	Job *Job
}

type JobSchedulePlan struct {
	Job *Job //任务信息
	Expr *cronexpr.Expression //解析好的表达式
	NextTime time.Time //下次调度时间
}

type JobExecuteInfo struct {
	Job *Job	//任务信息
	PlanTime time.Time	//理论调度时间
	RealTime time.Time  //实际调度时间
}

//任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output []byte
	Err error
	StartTime time.Time
	EndTime time.Time
}

/**
	反序列化json数据(仅用于job)
 */
func UnpackJob(value []byte) (ret *Job,err error){
	var job *Job
	job = &Job{}
	if err = json.Unmarshal(value,job);err != nil{
		return
	}
	ret = job
	return
}

/**
	获取etcd中的末尾内容
 */
func ExtractJobName(jobKey string) string{
	if jobKey == ""{
		return ""
	}
	return strings.TrimPrefix(jobKey,os.Getenv("ETCD_JOB_DIR"))
}

/**
	构建一个事件
 */
func BuildJobEvent(eventType int,job *Job) (jobEvent *JobEvent){
	return &JobEvent{
		EventType:eventType,
		Job:job,
	}
}

/**
	构造调度任务计划
 */
func BuildJobSchedulerPlan(job *Job)(jobSchedulePlan *JobSchedulePlan , err error){
	var(
		expr *cronexpr.Expression
	)
	if expr ,err = cronexpr.Parse(job.CronExpr);err !=nil{
		return
	}
	jobSchedulePlan = &JobSchedulePlan{
		Job:job,
		Expr:expr,
		NextTime:expr.Next(time.Now()),
	}
	return
}

/**
	构造执行状态信息
 */
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan)(jobExecuteInfo *JobExecuteInfo){
	jobExecuteInfo = &JobExecuteInfo{
		Job:jobSchedulePlan.Job,
		PlanTime:jobSchedulePlan.NextTime,
		RealTime:time.Now(),
	}
	return
}

/**
根据url获取执行的返回
*/
func GetHttpResponsed(sendUrl string) (bodyString []byte,err error){
	var(
		cli *goz.Request
		resp *goz.Response
		respBody goz.ResponseBody
	)
	cli = goz.NewClient()
	if resp , err = cli.Get(sendUrl,goz.Options{
		Headers: map[string]interface{}{
			"Accept": "application/json, text/javascript, */*; q=0.01",
			"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
			"Connection": "keep-alive",
			"Content-Type": "application/json",
			"Host": "www.baidu.com",
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36 Edg/80.0.361.111",
			"X-Requested-With": "XMLHttpRequest",
		},
	});err !=nil{
		logs.Info("请求时出错")
		return nil,err
	}
	if respBody ,err = resp.GetBody();err !=nil{
		logs.Info("获取body时出错")
		return nil,err
	}
	bodyString = respBody.Read(len(respBody.GetContents()))
	return
}
