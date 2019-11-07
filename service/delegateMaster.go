package service

import (
	"encoding/json"
	"net"
	"os"

	"watchmen/common"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hashicorp/memberlist"
)

type delegateMaster struct {
	broadcasts *memberlist.TransmitLimitedQueue
	extendIP   string
	peerlabel  string
	Reserved1  string
	logger     log.Logger
}

//delegateMaster工厂
func newDelegateMaster(m *memberlist.Memberlist, hostname string, extendIP string, webHookPort string) *delegateMaster {
	d := &delegateMaster{
		peerlabel: "master",
	}

	//http://cicd.yssredefinecloud.com/reg/updateKubeConfig
	if ip := net.ParseIP(extendIP); ip == nil {
		//空说明不是ip，按域名处理
		d.extendIP = "http://" + extendIP + "/reg/updateKubeConfig"
	} else {
		//否则按ip处理
		d.extendIP = "http://" + extendIP + ":" + webHookPort + "/updateKubeConfig"
	}

	//comma := strings.LastIndex(hostname, ":")
	//d.Reserved1 = "http://" + hostname[:comma] + ":" + webHookPort + "/echoSurvival"
	//d.Reserved1 = "http://" + hostname + ":" + webHookPort + "/echoSurvival"
	d.Reserved1 = "http://jenkins-ex-svc.devops.svc:" + webHookPort + "/echoSurvival"

	d.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	d.logger = log.NewLogfmtLogger(os.Stdout)
	d.logger = level.NewFilter(d.logger, Opt)

	return d

}

//memberlist delegate接口必须实现的方法(1/5)
func (d *delegateMaster) NodeMeta(limit int) []byte {
	//return []byte{}
	tmp := common.MateData{
		ExtendIP:  d.extendIP,
		Label:     d.peerlabel,
		Reserved1: d.Reserved1,
	}
	data, _ := json.Marshal(&tmp)
	return data
}

//memberlist delegate接口必须实现的方法(2/5)
func (d *delegateMaster) NotifyMsg(b []byte) {
}

//memberlist delegate接口必须实现的方法(3/5)
func (d *delegateMaster) GetBroadcasts(overhead, limit int) [][]byte {
	return d.broadcasts.GetBroadcasts(overhead, limit)
}

//memberlist delegate接口必须实现的方法(4/5)
//join为true是向所有节点push/pull，为false时定期随机选择一个节点push/pull
func (d *delegateMaster) LocalState(join bool) []byte {
	return []byte(d.extendIP)
}

//memberlist delegate接口必须实现的方法(5/5)
//接收到对端的LocalState消息会触发本方法
func (d *delegateMaster) MergeRemoteState(buf []byte, join bool) {
}

//memberlist EventDelegate接口必须实现的方法(1/3)
func (d *delegateMaster) NotifyJoin(n *memberlist.Node) {
	level.Debug(d.logger).Log("methodName", "NotifyJoin", "node", n.Name, "addr", n.Address(), "metaData", n.Meta)
	if len(n.Meta) == 0 {
		return
	}
	var unMateData common.MateData
	if err := json.Unmarshal(n.Meta, &unMateData); err != nil {
		level.Error(d.logger).Log("methodName", "NotifyJoin", "unMateData", err)
	}
	if unMateData.Label == "watchmen" {
		WatchmenOnline = true
	}
}

//memberlist EventDelegate接口必须实现的方法(2/3)
func (d *delegateMaster) NotifyLeave(n *memberlist.Node) {
	level.Debug(d.logger).Log("methodName", "NotifyLeave", "node", n.Name, "addr", n.Address(), "metaData", n.Meta)
	if len(n.Meta) == 0 {
		return
	}
	var unMateData common.MateData
	if err := json.Unmarshal(n.Meta, &unMateData); err != nil {
		level.Error(d.logger).Log("methodName", "NotifyLeave", "unMateData", err)
	}
	if unMateData.Label == "watchmen" {
		WatchmenOnline = false
	}
}

//memberlist EventDelegate接口必须实现的方法(3/3)
func (d *delegateMaster) NotifyUpdate(n *memberlist.Node) {
	level.Debug(d.logger).Log("methodName", "NotifyUpdate", "node", n.Name, "addr", n.Address(), "metaData", n.Meta)
}
