package main

import (
	"flag"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"watchmen/service"
	"watchmen/service/process"
)

var opt level.Option

var cliType = flag.String("clitype", "master", "client tyep")
var bindPort = flag.Int("bindPort", 7946, "gossip bindport")
var advertisePort = flag.Int("advertisePort", 7946, "gossip advertisePort")
var hostname = flag.String("peerName", "cicd-jenkins-srv-agent.devops.svc", "peer hostname")
var knownPeers = flag.String("knownPeers", "jenkins-watchmen.devops.svc:7946", "gossip need seed node")
var logLevel = flag.String("logLevel", "info", "output log level")

//master需要
var extendIP = flag.String("extendIP", "", "master extend ip or ingress domain name")

//slave需要
var peerlabel = flag.String("label", "", "peer meta data")
var gitlabURL = flag.String("gitlab_url", "", "kubeconfig repo gitlab url")
var token = flag.String("gitlab_token", "", "access gitlab need the token")

func main() {
	flag.Parse()

	if os.Getenv("LOG_LEVEL") != "" {
		*logLevel = os.Getenv("LOG_LEVEL")
	}

	setLogLevel(*logLevel)

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, opt)

	switch *cliType {
	case "master":
		fmt.Println("this is master")

		//使用client-go获取ingress中域名
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		ingress, err := client.ExtensionsV1beta1().Ingresses("devops").Get("devops-ingress", v1.GetOptions{})
		if err == nil && len(ingress.Spec.Rules) != 0 {
			var tmp string
			for _, rule := range ingress.Spec.Rules {
				tmp = rule.Host
				for _, path := range rule.HTTP.Paths {
					if path.Backend.ServiceName == "jenkins-ex-svc" {
						*extendIP = tmp
					}
				}
			}
		} else {
			level.Error(logger).Log("getIngressErr", err)
		}

		pr, err := service.CreatePeerMaster(*bindPort, *advertisePort, *hostname, *knownPeers, *extendIP)
		if err != nil {
			panic("init memberlist err==>" + err.Error())
		}
		pr.Join()

		go func() {
			defer func() {
				if err := recover(); err != nil {
					level.Error(logger).Log("goroutimePanicERROR", err)
				}
			}()
			service.CreateWebService(pr).NewWebService()
		}()
		//捕获系统退出信号，退出前执行Leave()安全退出
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-stop
		level.Debug(logger).Log("msg", "Receive system kill signal peer leave")
		pr.Leave()
		return
	case "slave":
		fmt.Println("this is slave")
		//peerSlave创建之前clone kubeConfig项目并增加cron任务，避免与notifyJoin冲突
		process.CreateProcesser().KubeConfigInit()
		pr, err := service.CreatePeerSlave(*peerlabel, *bindPort, *advertisePort, *hostname, *knownPeers, *gitlabURL, *token)
		if err != nil {
			panic("init memberlist err==>" + err.Error())
		}
		pr.Join()

		//捕获系统退出信号，退出前执行Leave()安全退出
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-stop
		level.Debug(logger).Log("msg", "Receive system kill signal peer leave")
		pr.Leave()
		return
	case "watchmen":
		fmt.Println("this is watchmen")
		pr, err := service.CreatePeerWatchmen(*bindPort, *advertisePort, *hostname, *knownPeers)
		if err != nil {
			panic("init memberlist err==>" + err.Error())
		}
		pr.Join()

		//捕获系统退出信号，退出前执行Leave()安全退出
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-stop
		level.Debug(logger).Log("msg", "Receive system kill signal peer leave")
		pr.Leave()
		return
	default:
		level.Error(logger).Log("clitype is", *cliType)
	}

}

func setLogLevel(logLevel string) error {
	switch logLevel {
	case "debug":
		opt = level.AllowDebug()
	case "info":
		opt = level.AllowInfo()
	case "warn":
		opt = level.AllowWarn()
	case "error":
		opt = level.AllowError()
	default:
		return fmt.Errorf("unrecognized log level %q", logLevel)
	}
	service.Opt = opt
	process.Opt = opt
	return nil
}
