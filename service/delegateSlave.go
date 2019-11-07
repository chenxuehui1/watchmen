package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hashicorp/memberlist"
	"watchmen/common"
	"watchmen/service/process"
)

type delegateSlave struct {
	broadcasts *memberlist.TransmitLimitedQueue
	ml         *memberlist.Memberlist
	gitlabURL  string
	token      string
	peerlabel  string
	extendIP   string
	online     chan int
	ch         chan string
	offline    chan int
	mark       bool
	logger     log.Logger
}

//delegateSlave工厂
func newDelegateSlave(m *memberlist.Memberlist, gitlabURL string, token string, peerlabel string) *delegateSlave {
	d := &delegateSlave{
		gitlabURL: gitlabURL,
		token:     token,
		peerlabel: peerlabel,
	}
	d.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	d.online = make(chan int, 1)
	d.ch = make(chan string, 1)
	d.offline = make(chan int, 1)
	d.logger = log.NewLogfmtLogger(os.Stdout)
	d.logger = level.NewFilter(d.logger, Opt)

	go d.worker()

	return d

}

//memberlist delegate接口必须实现的方法(1/5)
func (d *delegateSlave) NodeMeta(limit int) []byte {
	//return []byte{}
	tmp := common.MateData{
		ExtendIP: "",
		Label:    d.peerlabel,
	}
	data, _ := json.Marshal(&tmp)
	return data
}

//memberlist delegate接口必须实现的方法(2/5)
func (d *delegateSlave) NotifyMsg(b []byte) {
	msg, err := common.SetMsg(b)
	if err != nil {
		level.Error(d.logger).Log("methodName", "NotifyMsg", "getMsg :", err)
		return
	}
	level.Debug(d.logger).Log("methodName", "NotifyMsg", "ActionIs", (*msg).Action)
	switch (*msg).Action {
	case "NotifyUpdateKube":
		go func() {
			defer func() {
				if err := recover(); err != nil {
					level.Error(d.logger).Log("goroutimePanicERROR", err)
				}
			}()
			//收到notifyUpdatekube消息，更新kubeconifg
			process.CreateProcesser().UpdateKubeByCLI()
		}()
	default:
		level.Debug(d.logger).Log("methodName", "NotifyMsg", "msg", "into switch default")
	}
}

//memberlist delegate接口必须实现的方法(3/5)
func (d *delegateSlave) GetBroadcasts(overhead, limit int) [][]byte {
	return d.broadcasts.GetBroadcasts(overhead, limit)
}

//memberlist delegate接口必须实现的方法(4/5)
//join为true是向所有节点push/pull，为false时定期随机选择一个节点push/pull
func (d *delegateSlave) LocalState(join bool) []byte {
	return []byte(d.extendIP)
}

//memberlist delegate接口必须实现的方法(5/5)
//接收到对端的LocalState消息会触发本方法
func (d *delegateSlave) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	//master上线时的knownPeer不包括slave，开启join过滤后master后于slave上线将不能同步新数据到slave
	/*if !join {
		return
	}*/
	go func() {
		if d.extendIP == "" {
			d.extendIP = string(buf)
			d.ch <- string(buf)
		} else {
			d.extendIP = string(buf)
		}
		level.Debug(d.logger).Log("methodName", "MergeRemoteState", "slaveGetExtendIP", d.extendIP)
	}()
}

//memberlist EventDelegate接口必须实现的方法(1/3)
func (d *delegateSlave) NotifyJoin(n *memberlist.Node) {
	level.Debug(d.logger).Log("methodName", "NotifyJoin", "node", n.Name, "addr", n.Address(), "metaData", n.Meta)
	if len(n.Meta) == 0 {
		return
	}
	var unMateData common.MateData
	if err := json.Unmarshal(n.Meta, &unMateData); err != nil {
		level.Error(d.logger).Log("methodName", "NotifyJoin", "unMateData", err)
	}

	level.Debug(d.logger).Log("methodName", "NotifyJoin", "extendIP", unMateData.ExtendIP, "label", unMateData.Label)
	if unMateData.Label == "watchmen" {
		if d.extendIP != "" {
			level.Debug(d.logger).Log("methodName", "NotifyJoin", "msg", "d.extendIP exist value")
			d.online <- 1
			level.Debug(d.logger).Log("methodName", "NotifyJoin", "msg", "d.online <- 1")
		}
		d.mark = true
		level.Debug(d.logger).Log("methodName", "NotifyJoin", "d.mark", d.mark)

	}
}

//memberlist EventDelegate接口必须实现的方法(2/3)
func (d *delegateSlave) NotifyLeave(n *memberlist.Node) {
	level.Debug(d.logger).Log("methodName", "NotifyLeave", "node", n.Name, "addr", n.Address(), "metaData", n.Meta)
	if len(n.Meta) == 0 {
		return
	}
	var unMateData common.MateData
	if err := json.Unmarshal(n.Meta, &unMateData); err != nil {
		level.Error(d.logger).Log("methodName", "NotifyLeave", "unMateData", err)
	}
	if unMateData.Label == "watchmen" {
		d.offline <- 1
		level.Debug(d.logger).Log("methodName", "NotifyLeave", "msg", "d.offline <- 1")
		d.mark = false
		level.Debug(d.logger).Log("methodName", "NotifyLeave", "d.mark", d.mark)

	}
	if unMateData.Label == "master" {
		level.Debug(d.logger).Log("methodName", "NotifyLeave", "nodeRole", unMateData.Label)
		go d.findMaster(unMateData)
	}
}

//memberlist EventDelegate接口必须实现的方法(3/3)
func (d *delegateSlave) NotifyUpdate(n *memberlist.Node) {
	level.Debug(d.logger).Log("methodName", "NotifyUpdate", "node", n.Name, "addr", n.Address(), "metaData", n.Meta)
}

//借助协程管道顺序执行RegistyHook()和OpenCronJob()同时又不阻塞delegate中的方法
func (d *delegateSlave) worker() {
	defer func() {
		if err := recover(); err != nil {
			level.Error(d.logger).Log("goroutimePanicERROR", err)
		}
	}()
	for {
		select {
		case <-d.online:
			level.Debug(d.logger).Log("methodName", "worker", "msg", "watchmen online change Trigger mode")
			process.CreateProcesser().RegistyHook(d.gitlabURL, d.extendIP, d.token)
		case extendIP := <-d.ch:
			level.Debug(d.logger).Log("methodName", "worker", "msg", "get extendIP")
			if d.mark == true {
				level.Debug(d.logger).Log("methodName", "worker", "msg", "watchmen online get extendIP & change Trigger mode")
				process.CreateProcesser().RegistyHook(d.gitlabURL, extendIP, d.token)
			}

		case <-d.offline:
			level.Debug(d.logger).Log("methodName", "worker", "msg", "watchmen offline change Trigger mode")
			process.CreateProcesser().OpenCronJob()

		}

	}

}

func (d *delegateSlave) findMaster(unMateData common.MateData) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	stop := make(chan bool)
	for {
		select {
		case <-ticker.C:
			level.Debug(d.logger).Log("methodName", "findMaster", "msg", "<-ticker")
			go func() {
				defer func() {
					//捕获panic异常，避免程序退出
					if err := recover(); err != nil {
						level.Debug(d.logger).Log("methodName", "NotifyLeave", "panic", err)
						if v, ok := err.(string); ok == true {
							if v == "usePanicCloseApp" {
								//stop <- syscall.SIGKILL
								panic(errors.New("usePanicCloseApp"))
							}
						}
					}
				}()

				sendMSG := common.Message{
					Action: "IHopeYouAreHere",
					Data:   d.peerlabel,
				}
				data, err := json.Marshal(&sendMSG)
				resp, err := http.Post(unMateData.Reserved1, "application/json", bytes.NewReader(data))
				if err != nil {
					level.Info(d.logger).Log("echoSerialCall", err)
				}
				defer resp.Body.Close()
				b, _ := ioutil.ReadAll(resp.Body)

				var ReplyMsg struct {
					Msg string
				}
				json.Unmarshal(b, &ReplyMsg)
				level.Debug(d.logger).Log("echoSerialCallMSG", ReplyMsg.Msg)
				if ReplyMsg.Msg == "keep" {
					stop <- true
				}
				if ReplyMsg.Msg == "restart" {
					panic("usePanicCloseApp")
				}

			}()
		case <-stop:
			level.Debug(d.logger).Log("methodName", "findMaster", "msg", "<-stop")
			goto Exit
		}
	}
Exit:
	level.Debug(d.logger).Log("methodName", "findMaster", "msg", "goto exit")
}
