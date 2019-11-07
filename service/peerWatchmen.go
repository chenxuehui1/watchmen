package service

import (
	"os"
	"strings"

	"github.com/hashicorp/memberlist"
)

type peerWatchmen struct {
	peer
}

//CreatePeerWatchmen 为peerWatchmen工厂
func CreatePeerWatchmen(
	bindPort int,
	advertisePort int,
	hostname string,
	knownPeers string,
) (*peerWatchmen, error) {

	p := &peerWatchmen{}

	kp := os.Getenv("KNOWN_PEERS")
	if len(kp) > 0 {
		p.seedPeers = strings.Split(kp, ";")
	} else {
		p.seedPeers = strings.Split(knownPeers, ";")
	}

	if os.Getenv("PEER_NAME") != "" {
		hostname = os.Getenv("PEER_NAME")
	}

	delegate := newDelegateWatchmen(p.mlist)

	config := memberlist.DefaultLANConfig()
	config.Delegate = delegate
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
