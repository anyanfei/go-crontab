package service

import (
	"context"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/clientv3"
	"os"
)

type JobLock struct {
	//etcd客户端
	kv clientv3.KV
	lease clientv3.Lease
	leaseId clientv3.LeaseID
	cancelFunc context.CancelFunc
	jobName string //任务名
	isLocked bool

}

//初始化锁对象
func InitJobLock(jobName string,kv clientv3.KV,lease clientv3.Lease)(jobLock *JobLock){
	jobLock = &JobLock{
		kv:kv,
		lease: lease,
		jobName: jobName,
	}
	return
}

//尝试上锁
func (jobLock *JobLock)TryLock()(err error){
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
		leaseId clientv3.LeaseID
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
		txn clientv3.Txn
		lockKey string
		txnResp *clientv3.TxnResponse
	)
	//1、创建租约
	if leaseGrantResp , err = jobLock.lease.Grant(context.TODO(),5);err != nil{
		return err
	}
	//获取续租id
	leaseId = leaseGrantResp.ID
	//取消自动续租
	cancelCtx,cancelFunc = context.WithCancel(context.TODO())
	//2、自动续租
	if keepRespChan , err = jobLock.lease.KeepAlive(cancelCtx,leaseId);err !=nil{
		goto FAIL
	}
	//处理续租应答
	go func() {
		var(
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <- keepRespChan:
				if keepResp == nil{
					goto END
				}
			}
		}
		END:
	}()

	//3、创建事物txn
	txn = jobLock.kv.Txn(context.TODO())
	//获取锁路径
	lockKey = os.Getenv("ETCD_JOB_LOCK_DIR")+jobLock.jobName
	//4、事物抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey),"=",0)).
		Then(clientv3.OpPut(lockKey,"",clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	//提交事务
	if txnResp , err = txn.Commit();err !=nil{
		goto FAIL
	}
	//5、成功返回，失败释放租约
	if !txnResp.Succeeded{
		err = errors.New("锁已被占用")
		goto FAIL
	}
	//抢锁成功
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return

	FAIL:
		cancelFunc()//取消租约自动续租
		if _,err = jobLock.lease.Revoke(context.TODO(),leaseId);err !=nil{
			return err
		}
		return
}

/**
	释放分布式锁
 */
func (jobLock *JobLock)UnLock() {
	if jobLock.isLocked{
		var(
			err error
		)
		jobLock.cancelFunc()
		if _,err = jobLock.lease.Revoke(context.TODO(),jobLock.leaseId);err !=nil{
			logs.Error(err)
		}
	}
}