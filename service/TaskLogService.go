package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"go_crontab/models"
)

var(
	ormer orm.Ormer
	err error
)

/**
	获取日志列表
 */
func GetTaskLogsByPage(page, pageSize int, jobName string) map[string]interface{} {
	var(
		thisCount int64
		tl []models.TaskModel
	)
	var resultMap = make(map[string]interface{})
	ormer = orm.NewOrm()
	qs := ormer.QueryTable("xc_task_log")
	page = (page-1)*pageSize
	if _ , err = qs.Filter("job_name__icontains",jobName).OrderBy("-create_time").Limit(pageSize,page).All(&tl);err !=nil{
		logs.Error("查询时出错")
		logs.Error(err)
	}

	if thisCount ,err = qs.Filter("job_name__icontains",jobName).Count();err != nil{
		logs.Error("查询时出错")
		logs.Error(err)
	}
	resultMap["all_count"] = thisCount
	resultMap["lists"] = tl
	return resultMap
}
