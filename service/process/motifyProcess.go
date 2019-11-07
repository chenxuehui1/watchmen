package process

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

//Opt 为log level
var Opt level.Option

type processer struct {
	token     string
	gitlabURL string
	logger    log.Logger
}

//CreateProcesser 为processer工厂
func CreateProcesser() *processer {
	p := &processer{}
	p.logger = log.NewLogfmtLogger(os.Stdout)
	p.logger = level.NewFilter(p.logger, Opt)
	return p
}

func (p *processer) KubeConfigInit() {
	level.Debug(p.logger).Log("msg", "into KubeConfigInit")
	if _, err := exec.Command("/usr/local/bin/jenkinsGetKubeEnv2.sh", "-d").Output(); err != nil {
		level.Error(p.logger).Log("KubeConfigInitErr", err)
	}
}

func (p *processer) UpdateKubeByCLI() {
	level.Debug(p.logger).Log("msg", "into NotifyUpdateKube")
	if _, err := exec.Command("/usr/local/bin/jenkinsGetKubeEnv2.sh", "-i").Output(); err != nil {
		level.Error(p.logger).Log("UpdateKubeByCLIErr", err)
	}
}

func (p *processer) OpenCronJob() {
	level.Debug(p.logger).Log("msg", "into OpenCronJob Method")
	if _, err := exec.Command("/usr/local/bin/jenkinsGetKubeEnv2.sh", "-c").Output(); err != nil {
		level.Error(p.logger).Log("OpenCronJobErr", err)
	}
}

func (p *processer) cancelCronJob() {
	level.Debug(p.logger).Log("msg", "into cancelCronJob Method")
	if _, err := exec.Command("/usr/local/bin/jenkinsGetKubeEnv2.sh", "-e").Output(); err != nil {
		level.Error(p.logger).Log("cancelCronJobErr", err)
	}
}

func (p *processer) RegistyHook(gitlabURL, hookURL, token string) {
	level.Debug(p.logger).Log("msg", "into RegistyHook Method")
	gitlabIP, projectName, pathWithNamespace := p.getIPAndNameByURL(gitlabURL)
	projectID := p.getIDByProjectName(gitlabIP, token, projectName, pathWithNamespace)
	p.addProjectHookBytoken(gitlabIP, hookURL, token, projectID)
	p.cancelCronJob()
}

//Playload 为注册webhook的json结构体
type Playload struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	PushEvents bool   `json:"push_events"`
	NoteEvents bool   `json:"note_events"`
}

//Project 为gitlab返回的项目结构
type Project struct {
	ID                       int    `json:"id"`
	ProjectName              string `json:"name"`
	Namespace                string `json:"name_with_namespace"`
	ProjectPath              string `json:"path"`
	ProjectPathWithNamespace string `json:"path_with_namespace"`
	SSHURLToRepo             string `json:"ssh_url_to_repo"`
	HTTPURLToRepo            string `json:"http_url_to_repo"`
	WebURL                   string `json:"web_url"`
}

//Projects 为gitlab返回的项目数组
type Projects []Project

//HookList 为gitlab webhook结构体
type HookList struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	ProjectID int    `json:"project_id"`
}

//http://192.168.102.73:8081/chenhanming/testconfig.git
func (p *processer) getIPAndNameByURL(giturl string) (gitlabIP, projectName string, pathWithNamespace string) {
	url, err := url.Parse(giturl)
	if err != nil {
		level.Error(p.logger).Log("methodName", "getIPAndNameByURL", "giturlParseErr", err)
	}
	head := strings.LastIndex(giturl, "/")
	if head == -1 {
		level.Error(p.logger).Log("methodName", "getIPAndNameByURL", "Err", "giturl format illegal")
	}
	tail := strings.LastIndex(giturl, ".git")
	gitlabIP = url.Host
	//projectName所在namespace路径
	pathWithNamespace = url.Path
	projectName = giturl[head+1 : tail]
	level.Debug(p.logger).Log("methodName", "getIPAndNameByURL", "gitlabIP", gitlabIP, "projectName", projectName, "pathWithNamespace", url.Path)
	return

}

func (p *processer) getIDByProjectName(gitlabIP, token, projectName, pathWithNamespace string) (id string) {
	//http://192.168.102.73:8081/api/v4/projects?search=testconfig&private_token=9cnjLhc_L3tHe2y-N9nf
	var stringBuilder bytes.Buffer
	stringBuilder.WriteString("http://")
	stringBuilder.WriteString(gitlabIP)
	stringBuilder.WriteString("/api/v4/projects?search=")
	stringBuilder.WriteString(projectName)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", stringBuilder.String(), nil)
	//req.Header.Add("PRIVATE-TOKEN", "9cnjLhc_L3tHe2y-N9nf")
	req.Header.Add("PRIVATE-TOKEN", token)
	resp, err := client.Do(req)
	if err != nil {
		level.Error(p.logger).Log("methodName", "getIDByProjectName", "getProjectIdErr", err)
	}
	if resp.StatusCode != http.StatusOK {
		level.Error(p.logger).Log("methodName", "getIDByProjectName", "notifyReturnHttpStatusCodeErr", resp.StatusCode)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)

	var pjs Projects
	if err := json.Unmarshal(content, &pjs); err != nil {
		level.Error(p.logger).Log("methodName", "getIDByProjectName", "UnmarshalProjectsErr", err)
	}
	//不同命名空间下能够存在同名项目，因此需要比较namespace确定是哪个项目
	for _, v := range pjs {
		level.Debug(p.logger).Log("methodName", "getIDByProjectName", "ID", v.ID, "ProjectPathWithNamespace", v.ProjectPathWithNamespace, "pathWithNamespace", pathWithNamespace)
		if ("/" + v.ProjectPathWithNamespace + ".git") == pathWithNamespace {
			return strconv.Itoa(v.ID)
		}
	}
	level.Error(p.logger).Log("methodName", "getIDByProjectName", "Err", "notFindID")
	return ""
}

func (p *processer) addProjectHookBytoken(gitlabIP, hookURL, token string, projectID string) {

	pl := new(Playload)
	pl.ID = projectID
	pl.URL = hookURL
	pl.PushEvents = true
	pl.NoteEvents = true
	jsonPl, _ := json.Marshal(pl)
	level.Debug(p.logger).Log("jsonPl", jsonPl)

	var stringBuilder bytes.Buffer
	stringBuilder.WriteString("http://")
	stringBuilder.WriteString(gitlabIP)
	stringBuilder.WriteString("/api/v4/projects/")
	stringBuilder.WriteString(projectID)
	//stringBuilder.WriteString("/hooks?private_token=9cnjLhc_L3tHe2y-N9nf")
	stringBuilder.WriteString("/hooks?private_token=")
	stringBuilder.WriteString(token)

	client := &http.Client{}

	//获得该项目的所有webHook
	content, err := client.Get(stringBuilder.String())
	if err != nil {
		level.Error(p.logger).Log("methodName", "addProjectHookBytoken", "getProjectHookListErr", err)
		return
	}
	defer content.Body.Close()
	b, _ := ioutil.ReadAll(content.Body)

	hl := make([]HookList, 0)
	if err := json.Unmarshal(b, &hl); err != nil {
		level.Error(p.logger).Log("methodName", "addProjectHookBytoken", "UnmarshalHookListerr", err)
	}
	//判断要添加的webhook是否已存在
	for _, v := range hl {
		if v.URL == hookURL {
			level.Warn(p.logger).Log("methodName", "addProjectHookBytoken", "Warning", "add hook duplicate")
			level.Debug(p.logger).Log("methodName", "addProjectHookBytoken", "RegisteredHookUrl", v.URL, "hookURL", hookURL)
			return
		}
	}
	//添加webhook
	resp, err := client.Post(stringBuilder.String(), "application/json", bytes.NewReader(jsonPl))
	if err != nil {
		level.Error(p.logger).Log("methodName", "addProjectHookBytoken", "addProjectHookErr", err)
		return
	}
	defer resp.Body.Close()
	b, _ = ioutil.ReadAll(resp.Body)
	level.Debug(p.logger).Log("methodName", "addProjectHookBytoken", "addProjectHook", string(b))

}
