package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go_crontab/common"
	"os"
	"strconv"
	"strings"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}


var(
	G_jobMgr *JobMgr
)

func InitJobMgr()(err error){
	var(
		config clientv3.Config
		client *clientv3.Client
		kv	clientv3.KV
		lease	clientv3.Lease
		dialTimeoutInfo int64
		endPoints []string
		endPointsString string
	)


	endPointsString = os.Getenv("ETCD_END_POINTS")
	endPoints = strings.Split(endPointsString,",")
	dialTimeoutInfo, err = strconv.ParseInt(os.Getenv("ETCD_DIAL_TIMEOUT"), 10, 64)
	//初始化ETCD配置
	config = clientv3.Config{
		Endpoints:endPoints,
		DialTimeout:time.Duration(dialTimeoutInfo) * time.Millisecond,
	}
	//建立连接
	if client ,err = clientv3.New(config);err !=nil{
		return
	}

	//获取KV
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	G_jobMgr = &JobMgr{
		client:client,
		kv:kv,
		lease:lease,
	}
	return
}

//保存任务的方法
func(jobMgr *JobMgr) SaveJob(job *common.Job)(oldJob *common.Job,err error){
	var(
		jobKey string
		jobValue []byte
		putResp *clientv3.PutResponse
		oldJobObj common.Job
	)
	//设置etcd任务的key为:
	jobKey = os.Getenv("ETCD_JOB_DIR") + job.Name
	//设置etcd任务的value为：
	if jobValue , err = json.Marshal(job);err!=nil{
		return
	}
	//保存kv到etcd
	if putResp , err = jobMgr.kv.Put(context.TODO(),jobKey,string(jobValue),clientv3.WithPrevKV());err!=nil{
		return
	}
	if putResp.PrevKv != nil{
		if err = json.Unmarshal(putResp.PrevKv.Value,&oldJobObj);err!=nil{
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

//删除任务方法
func(jobMgr *JobMgr) DeleteJob(jobName string)(oldJob *common.Job,err error){
	var(
		jobKey string
		delResp *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	jobKey = os.Getenv("ETCD_JOB_DIR") + jobName
	if delResp , err = jobMgr.kv.Delete(context.TODO(),jobKey,clientv3.WithPrevKV());err !=nil{
		fmt.Println(err)
	}

	if len(delResp.PrevKvs) != 0{
		if err = json.Unmarshal(delResp.PrevKvs[0].Value,&oldJobObj);err!=nil{
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

//获取任务列表
func (jobMgr *JobMgr) GetListJob()(jobList []*common.Job,err error){
	var (
		jobKey string
		getResp *clientv3.GetResponse
		job *common.Job
	)
	jobKey = os.Getenv("ETCD_JOB_DIR")
	if getResp , err = jobMgr.kv.Get(context.TODO(),jobKey,clientv3.WithPrefix());err !=nil{
		return nil,err
	}
	jobList = make([]*common.Job,0)
	for _,getRespV := range getResp.Kvs{
		job = &common.Job{}
		if err = json.Unmarshal(getRespV.Value,job);err!=nil{
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return
}

/**
通过任务名杀死指定任务
*/
func (jobMgr *JobMgr) KillJob(name string)(err error){
	var(
		killerKey string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
	)

	killerKey = os.Getenv("ETCD_JOB_KILL_DIR") + name
	if leaseGrantResp,err = jobMgr.lease.Grant(context.TODO(),1);err !=nil{
		return
	}
	//获取到租约id
	leaseId = leaseGrantResp.ID
	if _,err = jobMgr.kv.Put(context.TODO(),killerKey,"",clientv3.WithLease(leaseId));err!=nil{
		return
	}
	return
}
