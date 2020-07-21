package common

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
		)
		result = &JobExecuteResult{
			ExecuteInfo:info,
			Output:make([]byte,0),
		}
		result.StartTime = time.Now()
		output,err = GetHttpResponsed(info.Job.Command)
		result.EndTime = time.Now()
		result.Output = output
		result.Err = err
		G_scheduler.PushJobResult(result)
	}()
}

//初始化执行器
func InitExecutor()(err error){
	G_executor = &Executor{}
	return
}
