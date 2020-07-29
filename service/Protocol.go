package service

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"os"
	"strings"
	"time"
)
//保存事件
const JOB_EVENT_SAVE int = 1
//删除事件
const JOB_EVENT_DELETE int = 2
//杀死事件
const JOB_EVENT_KILLER int = 3

type JobListRequest struct {
	Page int `json:"page"`
	PageSize int `json:"page_size"`
	KeyWord string `json:"key_word"`
}

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
	Job *Job                  //任务信息
	Expr *cronexpr.Expression //解析好的表达式
	NextTime time.Time        //下次调度时间
}

type JobExecuteInfo struct {
	Job *Job           //任务信息
	PlanTime time.Time //理论调度时间
	RealTime time.Time //实际调度时间
}

//任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output []byte
	Err error
	StartTime time.Time
	EndTime time.Time
}

//日志查询传入
type GetTaskLogs struct {
	Page int `json:"page"`
	PageSize int `json:"page_size"`
	JobName string `json:"job_name"`
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
 	从/cron/killer/job1中提取job1
 */
func ExtractKillerName(killerKey string) string{
	if killerKey == ""{
		return ""
	}
	return strings.TrimPrefix(killerKey,os.Getenv("ETCD_JOB_KILL_DIR"))
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
func BuildJobSchedulerPlan(job *Job)(jobSchedulePlan *JobSchedulePlan, err error){
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
	模拟php的array_chunk函数
 */
func SliceChunk(slice []interface{}, size int) (chunkslice [][]interface{}) {
	if size >= len(slice) {
		chunkslice = append(chunkslice, slice)
		return
	}
	var end int = size
	for i := 0; i <= len(slice); i += size {
		//再判断一下，若整个切片的长度都小于当前的末尾，则把末尾的值改为当前切片的最大值，否则越界
		if len(slice) < end {
			end = len(slice) + 1
		}
		chunkslice = append(chunkslice, slice[i:end])
		end += size
	}
	return
}

