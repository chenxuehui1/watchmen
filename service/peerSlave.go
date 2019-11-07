package service

import (
	"os"
	"strings"

	"github.com/hashicorp/memberlist"
)

type peerSlave struct {
	peer
}

//CreatePeerSlave 为peerSlave工厂
func CreatePeerSlave(
	peerlabel string,
	bindPort int,
	advertisePort int,
	hostname string,
	knownPeers string,
	gitlabURL string,
	token string) (*peerSlave, error) {

	p := &peerSlave{}

	if os.Getenv("GITLAB_URL") != "" {
		gitlabURL = os.Getenv("GITLAB_URL")
	}

	if os.Getenv("GITLAB_PWD") != "" {
		token = os.Getenv("GITLAB_PWD")
	}
	if os.Getenv("SWARM_CLIENT_LABELS") != "" {
		peerlabel = os.Getenv("SWARM_CLIENT_LABELS")
	}

	kp := os.Getenv("KNOWN_PEERS")
	if len(kp) > 0 {
		p.seedPeers = strings.Split(kp, ";")
	} else {
		p.seedPeers = strings.Split(knownPeers, ";")
	}

	delegate := newDelegateSlave(p.mlist, gitlabURL, token, peerlabel)

	config := memberlist.DefaultLANConfig()
	config.Delegate = delegate
	config.Events = delegate

	if hostname == "" {
		hostname, _ = os.Hostname()
	}
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
