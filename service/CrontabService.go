package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/gorhill/cronexpr"
	"time"
)

/**
	创建时间：2020-07-08 11:18:24
	根据传入的表达式，检查表达式是否正确：

      字段名       是否强制      支持的格式    		支持的特殊格式
	----------     ----------   --------------    --------------------
	Seconds        No           0-59              * / , -
	Minutes        Yes          0-59              * / , -
	Hours          Yes          0-23              * / , -
	Day of month   Yes          1-31              * / , - L W (L指最后last，若用L，表示月底,W指的是最近的工作日)
	Month          Yes          1-12 or JAN-DEC   * / , -
	Day of week    Yes          0-6 or SUN-SAT    * / , - L # (L指最后last，若用1L，表示本月最后一周的周一)
	Year           No           1970–2099         * / , -
 */
func CheckCrontabExpr(crontabs string) (err error,nextTimeArr []string){
	var nextTime []time.Time
	if _,err = cronexpr.Parse(crontabs);err !=nil{
		logs.Error(err)
		return err,nil
	}
	//返回当前crontab后的5次执行,n为次数
	nextTime = cronexpr.MustParse(crontabs).NextN(time.Now(), 5)
	for _,v := range nextTime{
		nextTimeArr = append(nextTimeArr,v.String())
	}
	return nil,nextTimeArr
}
