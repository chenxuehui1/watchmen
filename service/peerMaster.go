package service

import (
	"os"
	"strings"

	"github.com/hashicorp/memberlist"
)

type peerMaster struct {
	peer
}

//CreatePeerMaster 为peerMaster工厂
func CreatePeerMaster(
	bindPort int,
	advertisePort int,
	hostname string,
	knownPeers string,
	extendIP string,
) (*peerMaster, error) {

	p := &peerMaster{}

	kp := os.Getenv("KNOWN_PEERS")
	if len(kp) > 0 {
		p.seedPeers = strings.Split(kp, ";")
	} else {
		p.seedPeers = strings.Split(knownPeers, ";")
	}

	if os.Getenv("PEER_NAME") != "" {
		hostname = os.Getenv("PEER_NAME")
	}

	if os.Getenv("EXTEND_IP") != "" {
		extendIP = os.Getenv("EXTEND_IP")
	}

	var webHookPort string
	if os.Getenv("WEBHOOK_PORT") != "" {
		webHookPort = os.Getenv("WEBHOOK_PORT")
	} else {
		webHookPort = "8081"
	}

	delegate := newDelegateMaster(p.mlist, hostname, extendIP, webHookPort)

	config := memberlist.DefaultLANConfig()
	config.Delegate = delegate
	config.Events = delegate
	config.Name = hostname
	config.BindPort = bindPort
	if advertisePort != 0 {
		config.AdvertisePort = advertisePort
	} else {
		config.AdvertisePort = bindPort
	}

	ml, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	p.mlist = ml

	return p, nil

}

func (p *peerMaster) Status() []string {
	t := make([]string, len(p.mlist.Members()))
	for index, nodeMember := range p.mlist.Members() {
		t[index] = nodeMember.Name
	}
	return t
}

func (p *peerMaster) SendMsg(n *memberlist.Node, b []byte) error {
	return p.mlist.SendReliable(n, b)

}

func (p *peerMaster) GetNodeList() []*memberlist.Node {
	return p.mlist.Members()
}
