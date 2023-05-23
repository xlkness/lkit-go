package etcdv3

import (
	"context"
	"fmt"
	"github.com/xlkness/lkit-go/internal/log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpcxio/libkv/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const defaultTTL = 30

// EtcdConfigAutoSyncInterval give a choice to those etcd cluster could not auto sync
// such I deploy clusters in docker they will dial tcp: lookup etcd1: Try again, can just set this to zero
var EtcdConfigAutoSyncInterval = time.Minute * 5

// EtcdV3 is the receiver type for the Store interface
type EtcdV3 struct {
	session          *etcdV3Session
	AllowKeyNotFound bool
}

// Register registers etcd to libkv
//func guaranteeReplaceV3Store() {
//	libkv.AddStore(estore.ETCDV3, newV3Store)
//}

func newV3Store(addrs []string, options *store.Config) (store.Store, error) {
	client := &EtcdV3{}

	cfg := &clientv3.Config{
		Endpoints: addrs,
	}

	var invokeTime time.Duration
	if options != nil {
		invokeTime = options.ConnectionTimeout
		cfg.DialTimeout = options.ConnectionTimeout
		cfg.DialKeepAliveTimeout = options.ConnectionTimeout
		cfg.TLS = options.TLS
		cfg.Username = options.Username
		cfg.Password = options.Password

		cfg.AutoSyncInterval = EtcdConfigAutoSyncInterval
	}
	if invokeTime == 0 {
		invokeTime = 10 * time.Second
	}

	session, err := newEtcdV3Session(invokeTime, cfg)
	if err != nil {
		return nil, err
	}
	client.session = session
	log.Infof("use joymicro etcd v3 store for service discovery")
	return client, nil
}

// Put a value at the specified key
func (s *EtcdV3) Put(key string, value []byte, options *store.WriteOptions) error {
	var ttl int64
	if options != nil {
		ttl = int64(options.TTL.Seconds())
	}
	if ttl == 0 {
		ttl = defaultTTL
	}

	err := s.session.Put(key, value, int(ttl))
	if err != nil {
		return fmt.Errorf("put key(%v) error:%v", key, err)
	}
	return nil
}

// Get a value given its key
func (s *EtcdV3) Get(key string) (*store.KVPair, error) {
	return s.session.Get(key)
}

// Close closes the client connection
func (s *EtcdV3) Close() {
	s.session.Close()
}

func (s *EtcdV3) Delete(key string) error {
	return nil
}

func (s *EtcdV3) Exists(key string) (bool, error) {
	return false, nil
}

// Watch for changes on a key
func (s *EtcdV3) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return nil, nil
}

func (s *EtcdV3) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return nil, nil
}

func (s *EtcdV3) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return nil, nil
}

func (s *EtcdV3) List(directory string) ([]*store.KVPair, error) {
	return nil, nil
}

func (s *EtcdV3) DeleteTree(directory string) error {
	return nil
}

func (s *EtcdV3) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return false, nil, nil
}

func (s *EtcdV3) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return false, nil
}

type etcdV3Session struct {
	etcdClient struct {
		invokeTimeout time.Duration
		cfg           *clientv3.Config
		client        *clientv3.Client
		leaseID       clientv3.LeaseID
	}
	isSessionAlive int32
	mu             *sync.RWMutex
}

func newEtcdV3Session(invokeTimeout time.Duration, cfg *clientv3.Config) (*etcdV3Session, error) {
	session := &etcdV3Session{
		mu:             new(sync.RWMutex),
		isSessionAlive: 1,
	}

	session.etcdClient.invokeTimeout = invokeTimeout
	session.etcdClient.cfg = cfg

	return session, nil
}

func (session *etcdV3Session) recreateClient(entryPoint string, leaseTTL int) error {
	if session.etcdClient.client != nil {
		session.etcdClient.client.Close()
	}

	cli, err := clientv3.New(*session.etcdClient.cfg)
	if err != nil {
		return fmt.Errorf("创建配置(%+v)的etcd客户端报错:%v", session.etcdClient.cfg, err)
	}

	session.etcdClient.client = cli

	err = session.grant(leaseTTL)
	if err != nil {
		return err
	}

	log.Infof("entry(%v) recreate client ok, with ttl:%v, lease:%v", entryPoint, leaseTTL, session.etcdClient.leaseID)

	return nil
}

func (session *etcdV3Session) put(key string, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), session.etcdClient.invokeTimeout)
	_, err := session.etcdClient.client.Put(ctx, key, string(value), clientv3.WithLease(session.etcdClient.leaseID))
	cancel()
	if err == nil {
		err1 := session.keepAliveOnce()
		if err1 != nil {
			return fmt.Errorf("put key(%v) ok, and keepalive error:%v", key, err1)
		}
		return nil
	}
	return fmt.Errorf("put key(%v) error:%v", key, err)
}

func (session *etcdV3Session) keepAliveOnce() error {
	ctx, cancel := context.WithTimeout(context.Background(), session.etcdClient.invokeTimeout)
	resp, err := session.etcdClient.client.KeepAliveOnce(ctx, session.etcdClient.leaseID)
	cancel()
	if err != nil {
		return fmt.Errorf("keepalive once error:%v", err)
	}
	if resp == nil {
		return fmt.Errorf("keepalive once resp nil")
	}
	return nil
}

func (session *etcdV3Session) grant(ttl int) error {
	ctx, cancel := context.WithTimeout(context.Background(), session.etcdClient.invokeTimeout)
	resp, err := session.etcdClient.client.Grant(ctx, int64(ttl))
	cancel()
	if err != nil {
		return fmt.Errorf("new lease with ttl(%v) error:%v", ttl, err)
	}

	session.etcdClient.leaseID = resp.ID
	return nil
}

func (session *etcdV3Session) Get(key string) (*store.KVPair, error) {
	if atomic.LoadInt32(&session.isSessionAlive) != 1 {
		return nil, store.ErrCallNotSupported
	}

	ctx, cancel := context.WithTimeout(context.Background(), session.etcdClient.invokeTimeout)
	resp, err := session.etcdClient.client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, store.ErrKeyNotFound
	}

	pair := &store.KVPair{
		Key:       key,
		Value:     resp.Kvs[0].Value,
		LastIndex: uint64(resp.Kvs[0].Version),
	}

	return pair, nil
}

func (session *etcdV3Session) Put(key string, value []byte, ttl int) error {
	if atomic.LoadInt32(&session.isSessionAlive) != 1 {
		return store.ErrCallNotSupported
	}

	if session.etcdClient.client == nil {
		err := session.recreateClient("first_put", int(ttl))
		if err != nil {
			return err
		}
	}

	err := session.put(key, value)

	if err != nil && (strings.Contains(err.Error(), "grpc: the client connection is closing")) {
		// 链接断开，直接创建新客户端
		err := session.recreateClient("put", ttl)
		if err == nil {
			err1 := session.put(key, value)
			if err1 != nil {
				log.Warnf("put key(%v) error because connection is closing, and re put also error:%v", key, err1)
			} else {
				log.Infof("put key(%v) error because connection is closing, and re create client ok", key)
			}
			return err
		}
		log.Warnf("put key(%v) error because connection is closing, and re create client also error:%v", key, err)
		return err
	}

	if err != nil && strings.Contains(err.Error(), "requested lease not found") {
		err := session.grant(ttl)
		if err == nil {
			err1 := session.put(key, value)
			if err1 != nil {
				log.Warnf("put key(%v) error because lease not found, and re put also error:%v", key, err1)
			} else {
				log.Infof("put key(%v) error because lease not found, and re create lease ok", key)
			}
			return err
		}
		log.Warnf("put key(%v) error because lease not found, and re create lease also error:%v", key, err)
		return err
	}

	return err
}

func (session *etcdV3Session) Close() {
	if atomic.LoadInt32(&session.isSessionAlive) != 1 {
		return
	}
	atomic.StoreInt32(&session.isSessionAlive, 0)
}
