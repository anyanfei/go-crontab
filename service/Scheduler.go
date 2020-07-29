package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"go_crontab/models"
	"strconv"
	"time"
)

type Scheduler struct {
	JobEventChan chan *JobEvent                   //etcd任务事件队列
	JobPlanTable map[string] *JobSchedulePlan     //任务调度计划表
	JobExecutingTable map[string] *JobExecuteInfo //任务执行表
	JobResultChan chan *JobExecuteResult          //任务结果队列
}

var G_scheduler *Scheduler

/**
	调度任务
 */
func (scheduler *Scheduler) processJobEvent(jobEvent *JobEvent){
	var(
		jobSchedulerPlan *JobSchedulePlan
		err error
		jobExist bool
	)
	switch jobEvent.EventType {
		case JOB_EVENT_SAVE:
			if jobSchedulerPlan,err = BuildJobSchedulerPlan(jobEvent.Job);err !=nil{
				return
			}
			scheduler.JobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
		case JOB_EVENT_DELETE:
			if jobSchedulerPlan,jobExist = scheduler.JobPlanTable[jobEvent.Job.Name];jobExist{
				delete(scheduler.JobPlanTable,jobEvent.Job.Name)
			}
		case JOB_EVENT_KILLER://强杀任务事件
		//取消掉http请求执行

	}
}

/**
	删除正在表中的任务
 */
func (scheduler *Scheduler) processJobResult(result *JobExecuteResult){
	delete(scheduler.JobExecutingTable,result.ExecuteInfo.Job.Name)
	//在这里开始添加进入mysql日志xc_task_log
	var(
		outPutInfo string
		timeLayout string
		ormer orm.Ormer
		err error
		insertRows int64
		insertRowsString string

	)
	outPutInfo = string(result.Output)
	if outPutInfo == ""{
		outPutInfo = "接口返回空内容"
	}
	timeLayout = "2005-01-02 15:04:05"
	if outPutInfo != ""{
		taskLog := models.TaskModel{
			JobName: result.ExecuteInfo.Job.Name,
			JobRecallTime: result.EndTime,
			JobRecallContent: outPutInfo,
			CreateTime: time.Now(),
		}
		ormer = orm.NewOrm()
		if insertRows,err = ormer.Insert(&taskLog);err !=nil{
			logs.Error(err)
			logs.Error("写入日志时出错")
		}
		insertRowsString = strconv.FormatInt(insertRows,10)
		logs.Info("当前写入库日志id为："+insertRowsString)
	}
	logs.Info("任务执行完成",result.ExecuteInfo.Job.Name,result.Err,time.Unix(result.EndTime.Unix(),0).Format(timeLayout))
}

//协程启动调度
func (scheduler *Scheduler) schedulerRoutine(){
	var (
		jobEvent *JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobResult *JobExecuteResult
	)

	//初始化一次
	schedulerAfter = scheduler.beginScheduler()

	//使用原生定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	for{
		select {
			case jobEvent = <- scheduler.JobEventChan:
				scheduler.processJobEvent(jobEvent)
			case <- schedulerTimer.C:
			case jobResult = <- scheduler.JobResultChan:
				scheduler.processJobResult(jobResult)
			}
		//再调度一次任务
		schedulerAfter = scheduler.beginScheduler()
		schedulerTimer.Reset(schedulerAfter)
	}
}

//推送任务更改事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *JobEvent){
	scheduler.JobEventChan <- jobEvent
}

/**
	尝试执行任务
 */
func (scheduler *Scheduler) TryStartJob(jobPlan *JobSchedulePlan){
	var(
		jobExecuteInfo *JobExecuteInfo
		jobExecuting bool
	)
	if jobExecuteInfo,jobExecuting = scheduler.JobExecutingTable[jobPlan.Job.Name];jobExecuting{
		return
	}
	//构建执行状态
	jobExecuteInfo = BuildJobExecuteInfo(jobPlan)
	//保存执行状态
	scheduler.JobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo
	//执行任务
	logs.Info("执行任务：",jobExecuteInfo.Job.Name,"，执行时间：",jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

/**
	开始调度,重新计算任务调度状态
 */
func (scheduler *Scheduler) beginScheduler() (schedulerAfter time.Duration){
	//1.遍历所有任务
	var (
		jobPlan *JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)

	if len(scheduler.JobPlanTable) == 0{
		schedulerAfter = 1* time.Second
		return
	}

	now = time.Now()
	//2.检测到最近即将过期的任务
	for _,jobPlan = range scheduler.JobPlanTable{
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now){
			//执行任务
			scheduler.TryStartJob(jobPlan)
			//获取下一次执行时间
			/*logs.Info(jobPlan.Job.Name)
			logs.Info(jobPlan.NextTime)*/
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		//统计即将过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime){
			nearTime = &jobPlan.NextTime
		}
	}
	//计算当前时间的下一次调度间隔
	schedulerAfter = (*nearTime).Sub(now)
	return
}


//初始化调度器
func InitScheduler() (err error){
	go G_scheduler.schedulerRoutine()
	return
}

/**
	回传任务执行结果
 */
func (scheduler *Scheduler) PushJobResult(jobResult *JobExecuteResult)  {
	scheduler.JobResultChan <- jobResult
}