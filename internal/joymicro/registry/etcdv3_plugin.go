package registry

import (
	"fmt"
	"github.com/xlkness/lkit-go/internal/joymicro/registry/etcdv3"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/client"
	xclient "github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
)

// todo etcd插件不支持注册函数服务
func GetEtcdRegistryServerPlugin(key string, serviceAddr string, etcdAddress []string, hbInterval time.Duration) (server.Plugin, error) {
	if key == "" {
		key = "tcp"
	}

	if hbInterval < time.Second*3 {
		hbInterval = time.Second * 3
	}

	baseDir := getBaseDir()

	r := &etcdv3.EtcdV3RegisterPlugin{
		ServiceAddress: key + "@" + serviceAddr,
		EtcdServers:    etcdAddress,
		BasePath:       baseDir,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: hbInterval,
	}
	err := r.Start()
	return r, err
}

func GetEtcdRegistryClientPlugin(service string, etcdServerAddrs []string) xclient.ServiceDiscovery {
	d, err := client.NewEtcdV3Discovery(getBaseDir(), service, etcdServerAddrs, true, nil)
	if err != nil {
		panic(fmt.Errorf("NewEtcdV3Discovery error:%v", err))
	}
	return d
}

func getBaseDir() string {
	if NameSpace != "" {
		if NameSpace[len(NameSpace)-1] == '/' {
			if DefaultBaseDir[0] == '/' {
				return NameSpace + DefaultBaseDir[1:]
			}
			return NameSpace + DefaultBaseDir
		}
		if DefaultBaseDir[0] == '/' {
			return NameSpace + DefaultBaseDir
		}
		return NameSpace + "/" + DefaultBaseDir
	}
	return DefaultBaseDir
}
