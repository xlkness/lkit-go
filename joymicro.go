package lkit_go

import (
	"github.com/xlkness/lkit-go/internal/joymicro/joyclient"
	"github.com/xlkness/lkit-go/internal/joymicro/joyservice"
	"time"
)

type JoyService = joyservice.ServicesManager
type JoyClient = joyclient.Service

func NewService(listenAddr, exposeAddr string, etcdServerAddrs []string) (*JoyService, error) {
	return joyservice.New(listenAddr, exposeAddr, etcdServerAddrs)
}

func NewServiceWithKey(key string, listenAddr, exposeAddr string, etcdServerAddrs []string) (*JoyService, error) {
	return joyservice.NewWithKey(key, listenAddr, exposeAddr, etcdServerAddrs)
}

func New(service string, etcdServerAddrs []string, callTimeout time.Duration, isPermanentSocketLink bool) *JoyClient {
	return joyclient.New(service, etcdServerAddrs, callTimeout, isPermanentSocketLink)
}
