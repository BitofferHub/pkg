package discovery

import (
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var reg RegData

type RegData struct {
	Reg registry.Registrar
}

func NewRegistrar(addrs []string) registry.Registrar {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: addrs,
	})
	if err != nil {
		panic(err)
	}
	// new reg with etcd client
	rg := etcd.New(client)
	reg.Reg = rg
	return rg
}

func GetRegistrar() *RegData {
	return &reg
}
