package fserver

import (
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"net"
	"time"
	"github.com/viphxin/xingo/timer"
	"github.com/viphxin/xingo/iface"
)

func init() {
	utils.GlobalObject.Protoc = &fnet.Protocol{}
	// --------------------------------------------init log start
	utils.ReSettingLog()
	// --------------------------------------------init log end
}

type Server struct {
	Port    int
	MaxConn int
	GenNum  uint32
}

func NewServer() iface.Iserver {
	s := &Server{
		Port:    utils.GlobalObject.TcpPort,
		MaxConn: utils.GlobalObject.MaxConn,
	}
	return s
}

func (this *Server) handleConnection(conn *net.TCPConn) {
	this.GenNum += 1
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)
	// conn.SetDeadline(time.Now().Add(time.Minute * 2))
	//fconn := fnet.NewConnection(conn, this.GenNum, &fnet.Protocol{})
	fconn := fnet.NewConnection(conn, this.GenNum, utils.GlobalObject.Protoc)
	//fconn := fnet.NewConnection(conn, this.GenNum, &cluster.RpcServerProtocol{})
	fconn.Start()
}

func (this *Server) Start() {
	go func() {
		if utils.GlobalObject.IsUsePool{
			//init workpool
			utils.GlobalObject.Protoc.InitWorker(utils.GlobalObject.PoolSize)
		}

		ln, err := net.ListenTCP("tcp", &net.TCPAddr{
			Port: this.Port,
		})
		if err != nil {
			logger.Error(err)
		}
		logger.Info("start xingo server...")
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				logger.Error(err)
			}
			//max client exceed
			if fnet.ConnectionManager.Len() >= utils.GlobalObject.MaxConn{
				conn.Close()
			}else{
				go this.handleConnection(conn)
			}
		}
	}()
}

func (this *Server) Stop() {
	logger.Info("stop xingo server!!!")
}

func (this *Server) AddRouter(router interface{}) {
	logger.Info("AddRouter")
	fnet.MsgHandleObj.AddRouter(router)
}

func (this *Server) CallLater(durations time.Duration, f func(v ...interface{}), args ...interface{}){
	delayTask := timer.NewTimer(durations, f, args)
	delayTask.Run()
}

func (this *Server) CallWhen(ts string, f func(v ...interface{}), args ...interface{}){
	loc, err_loc := time.LoadLocation("Local")
	if err_loc != nil{
		logger.Error(err_loc)
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, loc)
	now := time.Now()
	//logger.Info(t)
	//logger.Info(now)
	//logger.Info(now.Before(t) == true)
	if err == nil{
		if now.Before(t){
			this.CallLater(t.Sub(now), f, args...)
		}else{
			logger.Error("CallWhen time before now")
		}
	}else{
		logger.Error(err)
	}
}

func (this *Server)CallLoop(durations time.Duration, f func(v ...interface{}), args ...interface{}){
	go func() {
		delayTask := timer.NewTimer(durations, f, args)
		for {
			time.Sleep(delayTask.GetDurations())
			delayTask.GetFunc().Call()
		}
	}()
}
