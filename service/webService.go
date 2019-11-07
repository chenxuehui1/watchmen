package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hashicorp/memberlist"
	"watchmen/common"
)

//Opt 为log level
var Opt level.Option

//WatchmenOnline watchmen是否上线标志
var WatchmenOnline bool

type webService struct {
	master *peerMaster
	logger log.Logger
}

//CreateWebService 为webService工厂
func CreateWebService(p *peerMaster) *webService {
	ws := &webService{}
	ws.master = p
	ws.logger = log.NewLogfmtLogger(os.Stdout)
	ws.logger = level.NewFilter(ws.logger, Opt)
	return ws
}

func (ws *webService) updateKubeConfigHandler(w http.ResponseWriter, r *http.Request) {

	if WatchmenOnline == false {
		return
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	level.Debug(ws.logger).Log("methodName", "updateKubeConfigHandler", "Body", b)

	var hook common.WebHook
	if err := json.Unmarshal(b, &hook); err != nil {
		level.Error(ws.logger).Log("methodName", "updateKubeConfigHandler", "Unmarshal", err)
	}
	level.Debug(ws.logger).Log("methodName", "updateKubeConfigHandler", "GitHTTPURL", hook.Repository.GitHTTPURL)

	ModifiedNodes := make([]string, len(hook.Commits[0].Modified))
	copy(ModifiedNodes, hook.Commits[0].Modified)
	if len(hook.Commits[0].Added) != 0 {
		ModifiedNodes = append(ModifiedNodes, hook.Commits[0].Added...)
	}
	level.Debug(ws.logger).Log("methodName", "updateKubeConfigHandler", "ModifiedNodes", fmt.Sprintf("%v", ModifiedNodes))

	ws.rawSend(ModifiedNodes)

	w.Header().Set("Content-Type", "application/json")
	replyMsg := &struct {
		msg string
	}{
		msg: "thank you tell me the events",
	}
	Res, _ := json.Marshal(replyMsg)
	w.Write(Res)
}

func (ws *webService) echoSurvivalHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	level.Debug(ws.logger).Log("methodName", "echoSurvivalHandler", "Body", b)

	var msg common.Message
	if err := json.Unmarshal(b, &msg); err != nil {
		level.Error(ws.logger).Log("methodName", "echoSurvivalHandler", "Unmarshal", err)
	}
	level.Debug(ws.logger).Log("methodName", "echoSurvivalHandler", "Action", msg.Action)
	if msg.Action == "IHopeYouAreHere" {
		exist := "restart"
		for _, nodeMember := range ws.master.GetNodeList() {
			if len(nodeMember.Meta) != 0 {
				var unMateData common.MateData
				if err := json.Unmarshal(nodeMember.Meta, &unMateData); err != nil {
					level.Error(ws.logger).Log("methodName", "echoSurvivalHandler", "unMateData", err)
				}
				if unMateData.Label == msg.Data {
					exist = "keep"
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		replyMsg := &struct {
			Msg string
		}{
			Msg: exist,
		}
		Res, _ := json.Marshal(replyMsg)
		w.Write(Res)

	}
}

func (ws *webService) cstatusHandler(w http.ResponseWriter, r *http.Request) {
	nodes := ws.master.GetNodeList()
	t := make([]string, len(nodes))
	for index, node := range nodes {
		t[index] = node.Name
	}
	w.Header().Set("Content-Type", "application/json")
	json, _ := json.Marshal(t)
	w.Write(json)
}

func (ws *webService) NewWebService() {
	http.HandleFunc("/cstatus", ws.cstatusHandler)
	http.HandleFunc("/echoSurvival", ws.echoSurvivalHandler)
	http.HandleFunc("/updateKubeConfig", ws.updateKubeConfigHandler)
	if err := http.ListenAndServe("0.0.0.0:8081", nil); err != nil {
		panic(err)
	}
}

func (ws *webService) rawSend(ModifiedNodes []string) {
	for index, str := range ModifiedNodes {
		comma := strings.LastIndex(str, "-config")
		//test-cluster-config ===> test-cluster
		ModifiedNodes[index] = str[:comma]
	}
	level.Debug(ws.logger).Log("methodName", "rawSend", "ModifiedNodes", fmt.Sprintf("%v", ModifiedNodes))

	for _, ModifiedMember := range ModifiedNodes {
		for _, nodeMember := range ws.master.GetNodeList() {
			if len(nodeMember.Meta) == 0 {
				level.Warn(ws.logger).Log("methodName", "rawSend", "Warning", "No metadata")
				//如果元数据不存在采用pod名字匹配
				if nodeMember.Name[14:17] == ModifiedMember {
					ws.coreSend(nodeMember)
				}

			} else {
				var unMateData common.MateData
				if err := json.Unmarshal(nodeMember.Meta, &unMateData); err != nil {
					level.Error(ws.logger).Log("methodName", "rawSend", "unMateData", err)
				}
				//采用peer元数据匹配
				if unMateData.Label == ModifiedMember {
					ws.coreSend(nodeMember)
				}
			}

		}
	}
}

func (ws *webService) coreSend(nodeMember *memberlist.Node) {
	msg := common.Message{
		Action: "NotifyUpdateKube",
		Data:   "",
	}
	data, err := json.Marshal(&msg)
	if err != nil {
		level.Error(ws.logger).Log("methodName", "coreSend", "MarshalErr", err)
	}
	data = append([]byte("p"), data...)
	ws.master.SendMsg(nodeMember, data)
}
