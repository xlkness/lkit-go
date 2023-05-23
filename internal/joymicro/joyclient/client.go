package joyclient

import (
	"context"
	"github.com/xlkness/lkit-go/internal/joymicro/registry"
	"github.com/xlkness/lkit-go/internal/joymicro/rpc_tracer"
	"github.com/xlkness/lkit-go/internal/joymicro/util"
	"sync"
	"time"

	rotel "github.com/rpcxio/rpcx-plugins/client/otel"
	"github.com/smallnest/rpcx/client"
)

type Service struct {
	ServiceName string
	etcdAddrs   []string
	// 阻塞调用超时时间
	callTimeout time.Duration
	// 默认跟rpc server不是永久连接，如果实时通讯量大的话，设置为true，
	// 字段作用：如果跟server读超时就关闭socket连接，等待之后的请求重新connect，
	// 用来避免长链接，有通信需求的双方节点形成强联通图，无用established套接字太多
	isPermanentSocketLink bool
	client                client.XClient
	selector              client.Selector
	plugins               client.PluginContainer
	peerServicesLock      *sync.Mutex
	once                  *sync.Once
}

// New 创建对某个节点的rpc客户端管理结构
// etcdServerAddrs:etcd服务的多个节点地址
// callTimeout:调用服务超时时间
// isPermanentSocketLink:默认跟rpc server不是永久连接，如果实时通讯量大的话，设置为true，
//
//	字段作用：如果跟server读超时就关闭socket连接，等待之后的请求重新connect，
//	用来避免长链接，有通信需求的双方节点形成强联通图，无用established套接字太多
func New(service string, etcdServerAddrs []string, callTimeout time.Duration, isPermanentSocketLink bool) *Service {
	etcdServerAddrs = util.PreHandleEtcdHttpAddrs(etcdServerAddrs)

	c := &Service{
		ServiceName:           service,
		etcdAddrs:             etcdServerAddrs,
		callTimeout:           callTimeout,
		isPermanentSocketLink: isPermanentSocketLink,
		peerServicesLock:      &sync.Mutex{},
		once:                  new(sync.Once),
		plugins:               client.NewPluginContainer(),
	}

	return c
}

func (s *Service) SetSelector(selector client.Selector) {
	s.selector = selector
	if s.client != nil {
		s.client.SetSelector(selector)
	}
}

func (s *Service) enableTracer() {
	tp := rpc_tracer.GetJaegerTracerProvider()
	if tp == nil {
		return
	}

	tc := tp.Tracer(rpc_tracer.TracerName)
	p := rotel.NewOpenTelemetryPlugin(tc, nil)
	s.client.GetPlugins().Add(p)
}

// Call 根据负载算法从服务中挑一个调用
func (s *Service) Call(ctx context.Context, method string, args interface{}, reply interface{}) error {
	if _, find := ctx.Deadline(); !find {
		newCtx, f := context.WithTimeout(ctx, s.callTimeout)
		defer f()
		ctx = newCtx
	}
	c := s.getXClient()
	return c.Call(ctx, method, args, reply)
}

/*
	以下为脱离服务概念的接口
*/

// CallAll 调用所有节点，有一个调用返回错误，整个调用都错误
func (s *Service) CallAll(ctx context.Context, method string, args interface{}, reply interface{}) error {
	if _, find := ctx.Deadline(); !find {
		newCtx, f := context.WithTimeout(ctx, s.callTimeout*2)
		defer f()
		ctx = newCtx
	}
	c := s.getXClient()
	return c.Broadcast(ctx, method, args, reply)
}

func (s *Service) getXClient() client.XClient {
	if s.client != nil {
		return s.client
	}

	s.once.Do(s.newXClient)

	return s.client
}

func (s *Service) newXClient() {
	conf := client.DefaultOption
	conf.Retries = 4

	// 默认维持2min连接，读超时就和服务器断开链接
	conf.IdleTimeout = time.Minute * 2
	if s.isPermanentSocketLink {
		conf.Heartbeat = true
		conf.HeartbeatInterval = time.Second * 30
	}

	// conf.ReadTimeout = time.Second * 10
	// conf.WriteTimeout = time.Second * 10

	d := registry.GetEtcdRegistryClientPlugin(s.ServiceName, s.etcdAddrs)
	xclient := client.NewXClient(s.ServiceName, client.Failover, client.RandomSelect, d, conf)
	if s.selector != nil {
		xclient.SetSelector(s.selector)
	}
	if s.plugins != nil {
		xclient.SetPlugins(s.plugins)
	}
	s.client = xclient
	s.enableTracer()
}
