package service

import (
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/memberlist"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type peer struct {
	mlist     *memberlist.Memberlist
	seedPeers []string
}

//Join 将Peer加入集群
func (p *peer) Join() error {

	if _, err := p.mlist.Join(p.seedPeers); err != nil {
		return err
	}
	return nil
}

func (p *peer) Leave() {
	if err := p.mlist.Leave(time.Second * 10); err != nil {
		panic(err)
	}
}

func init() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8082", nil))
}
