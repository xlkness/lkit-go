package lkit_go

import (
	"github.com/smallnest/rpcx/client"
	"github.com/xlkness/lkit-go/internal/joymicro/joyclient"
	"github.com/xlkness/lkit-go/internal/joymicro/joyservice"
	"time"
)

type JoyService = joyservice.ServicesManager
type JoyClient = joyclient.Service

func NewRpcService(listenAddr, exposeAddr string, etcdServerAddrs []string) (*JoyService, error) {
	return joyservice.New(listenAddr, exposeAddr, etcdServerAddrs)
}

func NewRpcServiceWithKey(key string, listenAddr, exposeAddr string, etcdServerAddrs []string) (*JoyService, error) {
	return joyservice.NewWithKey(key, listenAddr, exposeAddr, etcdServerAddrs)
}

func NewRpcClient(service string, etcdServerAddrs []string, callTimeout time.Duration, isPermanentSocketLink bool) *JoyClient {
	return joyclient.New(service, etcdServerAddrs, callTimeout, isPermanentSocketLink)
}

func NewRpcConsistentHashSelector() client.Selector {
	return joyclient.NewConsistentHashSelector()
}

func NewRpcPeerSelector() client.Selector {
	return joyclient.NewPeerSelector()
}
