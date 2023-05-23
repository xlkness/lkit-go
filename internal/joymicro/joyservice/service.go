package joyservice

import (
	"context"
	"github.com/xlkness/lkit-go/internal/joymicro/registry"
	"github.com/xlkness/lkit-go/internal/joymicro/rpc_tracer"
	"github.com/xlkness/lkit-go/internal/joymicro/util"
	"net/url"
	"time"

	rotel "github.com/rpcxio/rpcx-plugins/client/otel"
	"github.com/smallnest/rpcx/server"
)

var DefaultEtcdHeartBeatInterval = time.Second * 3

type ServicesManager struct {
	ListenAddr string
	Addr       string         // 节点提供rpc服务的地址
	rpcserver  *server.Server // rpc服务器，用来注册服务，服务发现等
	isRunning  bool
}

// New 创建一个服务
// service:服务器名称
// addr:当前节点服务对外可以访问的地址，不是监听地址，必须为"ip+:+port"格式
// etcdServerAddrs:etcd服务地址
func New(listenAddr, exposeAddr string, etcdServerAddrs []string) (*ServicesManager, error) {
	return NewWithKey("", listenAddr, exposeAddr, etcdServerAddrs)
}

// NewWithKey 创建一个带主键的服务，用于点对点通信
// service:服务器名称
// addr:当前节点服务对外可以访问的地址，不是监听地址，必须为"ip+:+port"格式
// etcdServerAddrs:etcd服务地址
func NewWithKey(key string, listenAddr, exposeAddr string, etcdServerAddrs []string) (*ServicesManager, error) {
	etcdServerAddrs = util.PreHandleEtcdHttpAddrs(etcdServerAddrs)
	m := newServersManager(listenAddr, exposeAddr)

	// 添加etcd注册中心
	r, err := registry.GetEtcdRegistryServerPlugin(key, m.Addr, etcdServerAddrs, DefaultEtcdHeartBeatInterval)
	if err != nil {
		return nil, err
	}
	m.rpcserver.Plugins.Add(r)
	m.enableTracer()

	return m, nil
}

func (m *ServicesManager) enableTracer() {

	tp := rpc_tracer.GetJaegerTracerProvider()
	if tp == nil {
		return
	}

	tc := tp.Tracer(rpc_tracer.TracerName)

	p := rotel.NewOpenTelemetryPlugin(tc, nil)
	m.rpcserver.Plugins.Add(p)
}

// RegisterOneService 注册一个服务
// service：服务名
// handler：回调处理
func (m *ServicesManager) RegisterOneService(service string, handler interface{}, metaKVs map[string]string) error {
	if m == nil {
		return nil
	}

	values := make(url.Values)
	for k, v := range metaKVs {
		values.Add(k, v)
	}
	return m.rpcserver.RegisterName(service, handler, values.Encode())
}

// Run 启动rpc服务
// addr：监听地址，可以忽略ip，例如":8888"格式
// 注意：register过程必须在start之前
func (m *ServicesManager) Run() error {
	if m == nil {
		return nil
	}

	if m.isRunning {
		return nil
	}
	m.isRunning = true
	err := m.rpcserver.Serve("tcp", m.ListenAddr)
	if err == server.ErrServerClosed {
		err = nil
	}
	m.isRunning = false
	return err
}

func (m *ServicesManager) Stop() {
	if m == nil {
		return
	}
	if !m.isRunning {
		return
	}
	m.isRunning = false
	ctx, f := context.WithTimeout(context.Background(), time.Second*5)
	defer f()
	m.rpcserver.Shutdown(ctx)
}

func newServersManager(listenAddr, exposeAddr string) *ServicesManager {
	m := &ServicesManager{
		ListenAddr: listenAddr,
		Addr:       exposeAddr,
		rpcserver:  server.NewServer(),
	}

	return m
}
