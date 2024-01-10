package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	//"go.etcd.io/etcd/clientv3"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type DisData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Metadata struct {
	} `json:"metadata"`
	Endpoints []string `json:"endpoints"`
}

type ServicesDiscovery struct {
	ServiceDiscoveryList map[string]*ServiceDiscovery
}

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	cli        *clientv3.Client  //etcd client
	serverList map[string]string //服务列表
	lock       sync.Mutex
}

var dis *ServicesDiscovery
var cli *clientv3.Client

func InitServiceDiscovery(endpoints []string, serviceList []string) {
	tmpcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	cli = tmpcli
	if err != nil {
		log.Fatal(err)
	}
	dis = new(ServicesDiscovery)
	dis.ServiceDiscoveryList = make(map[string]*ServiceDiscovery, 0)
	//ser.WatchService("/microservices/user-svr")
	for _, serviceName := range serviceList {
		ser := NewServiceDiscovery(cli)
		//defer ser.Close()
		prefix := fmt.Sprintf("/microservices/%s", serviceName)
		fmt.Println("prefix:", prefix)
		ser.WatchService(prefix)
		dis.ServiceDiscoveryList[serviceName] = ser
	}
}

func GetServiceDiscovery(serviceName string) *ServiceDiscovery {
	return dis.ServiceDiscoveryList[serviceName]
}

// NewServiceDiscovery  新建发现服务
func NewServiceDiscovery(cli *clientv3.Client) *ServiceDiscovery {

	return &ServiceDiscovery{
		cli:        cli,
		serverList: make(map[string]string),
	}
}

// WatchService 初始化服务列表和监视
func (s *ServiceDiscovery) WatchService(prefix string) error {
	//根据前缀获取现有的key
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}

	//监视前缀，修改变更的server
	go s.watcher(prefix)
	return nil
}

// watcher 监听前缀
func (s *ServiceDiscovery) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: //删除
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

const (
	PUT    = 0
	DELETE = 1
	EXPIRE = 2
)

// SetServiceList 新增服务地址
func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.serverList[key] = string(val)
	log.Println("put key :", key, "val:", val)
}

// DelServiceList 删除服务地址
func (s *ServiceDiscovery) DelServiceList(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.serverList, key)
	log.Println("del key:", key)
}

// GetServices 获取服务地址
func (s *ServiceDiscovery) GetServices() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	addrs := make([]string, 0)

	for _, v := range s.serverList {
		addrs = append(addrs, v)
	}
	return addrs
}

// GetServices 获取服务地址
func (s *ServiceDiscovery) GetEndPoint() []string {
	addrs := s.GetServices()
	if len(addrs) == 0 {
		return []string{}
	}
	var disList = new(DisData)
	rand.Seed(time.Now().Unix())
	fmt.Println("addrs", addrs)
	pos := rand.Int63n(int64(len(addrs)))
	//		pos := rand.Intn(2)
	if pos > 0 {
		fmt.Println(true)
	}
	fmt.Println("pos ", pos)

	json.Unmarshal([]byte(addrs[pos]), disList)

	return disList.Endpoints
}

// GetHttpEndPoint 获取服务地址
func (s *ServiceDiscovery) GetHttpEndPoint() string {
	addrs := s.GetServices()
	if len(addrs) == 0 {
		return ""
	}
	var disList = new(DisData)
	rand.Seed(time.Now().Unix())
	pos := rand.Int63n(int64(len(addrs)))
	//		pos := rand.Intn(2)
	if pos > 0 {
		fmt.Println(true)
	}
	fmt.Println("pos ", pos)

	json.Unmarshal([]byte(addrs[pos]), disList)
	for _, endpoint := range disList.Endpoints {
		if strings.Contains(endpoint, "http") {
			return endpoint
		}
	}
	return ""
}

// GetGrpcEndPoint 获取服务地址
func (s *ServiceDiscovery) GetGrpcEndPoint() string {
	addrs := s.GetServices()
	if len(addrs) == 0 {
		return ""
	}
	var disList = new(DisData)
	rand.Seed(time.Now().Unix())
	pos := rand.Int63n(int64(len(addrs)))
	//		pos := rand.Intn(2)
	if pos > 0 {
		fmt.Println(true)
	}
	fmt.Println("pos ", pos)

	json.Unmarshal([]byte(addrs[pos]), disList)
	for _, endpoint := range disList.Endpoints {
		if strings.Contains(endpoint, "grpc") {
			return strings.Trim(endpoint, "grpc://")
		}
	}
	return ""
}

// Close 关闭服务
func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}
