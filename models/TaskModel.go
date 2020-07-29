package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//日志数据库
type TaskModel struct {
	Id int `json:"id";orm:"column(id)";orm:"PK"`
	JobName string `json:"job_name";orm:"column(job_name)"`
	JobRecallTime time.Time `json:"job_recall_time";orm:"column(job_recall_time)"`
	JobRecallContent string `json:"job_recall_content";orm:"column(job_recall_content)"`
	CreateTime time.Time `json:"create_time";orm:"column(create_time)"`
}

func init(){
	orm.RegisterModelWithPrefix("xc_",new(TaskModel))
}

func (m *TaskModel)TableName() string{
	return "task_log"
}