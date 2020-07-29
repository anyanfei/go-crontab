package service

import (
	"time"
)

type Executor struct {
}

var(
	G_executor *Executor
)

//执行一个任务
func (executor *Executor) ExecuteJob(info *JobExecuteInfo){
	go func() {
		var(
			err error
			output []byte
			result *JobExecuteResult
			jobLock *JobLock
		)
		result = &JobExecuteResult{
			ExecuteInfo:info,
			Output:make([]byte,0),
		}

		//初始化分布式锁
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)

		result.StartTime = time.Now()
		//上锁
		err = jobLock.TryLock()
		//出错直接释放锁
		defer jobLock.UnLock()
		//开始上锁
		if err !=nil{
			result.Err = err
			result.EndTime=time.Now()
		}else{
			//上锁成功后重置任务启动时间
			result.StartTime = time.Now()
			output,err = GetHttpResponsed(info.Job.Command)
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		G_scheduler.PushJobResult(result)
	}()
}

//初始化执行器
func InitExecutor()(err error){
	G_executor = &Executor{}
	return
}
