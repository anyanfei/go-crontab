package service

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"os"
	"time"
)

func InitDataBase()(err error){
	//mysql操作
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbDataBase := os.Getenv("DB_DATABASE")
	timeZone := "Asia/Shanghai"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s",dbUser,dbPass,dbHost,dbPort,dbDataBase,url.QueryEscape(timeZone))
	err = orm.RegisterDataBase("default","mysql",dsn)
	if err !=nil{
		return err
	}
	orm.SetMaxOpenConns("default",30)
	orm.SetMaxIdleConns("default",5)
	db,_:=orm.GetDB("default")
	db.SetConnMaxLifetime(3600 * time.Second)
	return
}
