package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/clientv3"
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
	watcher clientv3.Watcher
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
		watcher clientv3.Watcher
		dialTimeoutInfo int64
		endPoints []string
		endPointsString string
	)

	//顺带初始化调度器的内存空间，因为在这里就需要向通道推送数据了
	G_scheduler = &Scheduler{
		JobEventChan:make(chan *JobEvent,1000),
		JobPlanTable: make(map[string] *JobSchedulePlan),
		JobExecutingTable:make(map[string] *JobExecuteInfo),
		JobResultChan:make(chan *JobExecuteResult,1000),
	}


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
	watcher = clientv3.Watcher(client)

	G_jobMgr = &JobMgr{
		client:client,
		kv:kv,
		lease:lease,
		watcher:watcher,
	}
	//启动监听器任务
	if err = G_jobMgr.watchJobs();err !=nil{
		return
	}
	//启动监听强杀任务
	//G_jobMgr.watchKiller()

	return
}

//保存任务的方法
func(jobMgr *JobMgr) SaveJob(job *Job)(oldJob *Job,err error){
	var(
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj Job
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
func(jobMgr *JobMgr) DeleteJob(jobName string)(oldJob *Job,err error){
	var(
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj Job
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
func (jobMgr *JobMgr) GetListJob()(jobList []*Job,err error){
	var (
		jobKey string
		getResp *clientv3.GetResponse
		job *Job
	)
	jobKey = os.Getenv("ETCD_JOB_DIR")
	if getResp , err = jobMgr.kv.Get(context.TODO(),jobKey,clientv3.WithPrefix());err !=nil{
		return nil,err
	}
	jobList = make([]*Job,0)
	for _,getRespV := range getResp.Kvs{
		job = &Job{}
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

/**
	监听job任务的变化
 */
func (jobMgr *JobMgr) watchJobs() (err error){
	var(
		getResp *clientv3.GetResponse
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		job *Job
		jobName string
		jobEvent *JobEvent
	)
	if getResp , err = jobMgr.kv.Get(context.TODO(),os.Getenv("ETCD_JOB_DIR"),clientv3.WithPrefix());err != nil{
		return
	}

	//得到当前有哪些任务
	for _,respV := range getResp.Kvs{
		if job,err = UnpackJob(respV.Value);err == nil{
			jobEvent = BuildJobEvent(JOB_EVENT_SAVE,job)
			G_scheduler.PushJobEvent(jobEvent)
		}
	}

	//监听版本变化
	go func() {
		watchStartRevision = getResp.Header.Revision + 1
		//监听目录的变化，从watchStartRevision版本开始监听
		watchChan = jobMgr.watcher.Watch(context.TODO(),os.Getenv("ETCD_JOB_DIR"),clientv3.WithRev(watchStartRevision),clientv3.WithPrefix())
		//处理监听
		for watchResp = range watchChan{
			for _,watchEvent = range watchResp.Events{
				switch watchEvent.Type {
				case mvccpb.PUT:
					if job , err = UnpackJob(watchEvent.Kv.Value);err !=nil{
						continue
					}
					jobEvent = BuildJobEvent(JOB_EVENT_SAVE,job)
				case mvccpb.DELETE:
					jobName = ExtractJobName(string(watchEvent.Kv.Key))
					job = &Job{Name: jobName}
					jobEvent = BuildJobEvent(JOB_EVENT_DELETE,job)
				}
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()

	return
}

/**
	监听手动杀死任务的事件
 */
/*func (jobMgr *JobMgr) watchKiller(){
	var (
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent *JobEvent
		jobName string
		job *Job
	)
	//监听/cron/killer目录
	go func() {
		//监听协程/cron/killer的变化
		watchChan = jobMgr.watcher.Watch(context.TODO(),os.Getenv("ETCD_JOB_KILL_DIR"),clientv3.WithPrefix())
		//处理监听
		for watchResp = range watchChan{
			for _,watchEvent = range watchResp.Events{
				switch watchEvent.Type {
				case mvccpb.PUT://杀死任务事件
					jobName = ExtractKillerName(string(watchEvent.Kv.Key))//获取任务最后的名字
					job = &Job{Name: jobName}
					jobEvent = BuildJobEvent(JOB_EVENT_KILLER,job)
					//推送给push scheduler
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE://killer标记过期，被自动删除
				}
			}
		}
	}()
}*/

/**
	创建任务执行锁
 */

func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock){
	//返回一把锁
	jobLock = InitJobLock(jobName,jobMgr.kv,jobMgr.lease)
	return jobLock
}
