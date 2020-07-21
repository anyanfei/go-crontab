package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/idoubi/goz"
)

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
