package common

import (
	"encoding/json"
	"errors"
)

//MateData 为watchmen节点元数据
type MateData struct {
	ExtendIP  string `json:"extend_ip"`
	Label     string `json:"label"`
	Reserved1 string `json:"reserved_1"`
}

//Message 为发送消息体
type Message struct {
	Action string
	Data   string
}

type NotifyUpdateKube struct {
	PeerName string
}

func SetMsg(b []byte) (*Message, error) {
	if len(b) == 0 {
		return nil, errors.New("msg len is 0")
	}
	if b[0] != 'p' {
		return nil, errors.New("msg Type mismatch")
	}
	m := &Message{}
	if err := json.Unmarshal(b[1:], m); err != nil {
		return nil, errors.New("msg data Unmarshal err")
	}
	return m, nil
}
