package etcdv3

import (
	"context"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"testing"
	"time"
)

type Req struct {
	F int
}
type Res struct {
	F int
}

type H struct {
}

func (h *H) Hello(ctx context.Context, req *Req, res *Res) error {
	return nil
}

func wrapServer() {
	r := &EtcdV3RegisterPlugin{
		ServiceAddress: "tcp" + "@" + "0.0.0.0:8888",
		EtcdServers:    []string{"192.168.1.22:2415"},
		BasePath:       "/likun",
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Second * 3,
	}
	err := r.Start()
	if err != nil {
		panic(err)
	}
	r.Register("shop", nil, "1:2")
	fmt.Printf("start server...\n")
	select {}
}

func TestTick(t *testing.T) {
	wrapServer()
}
