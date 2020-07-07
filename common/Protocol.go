package common

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
