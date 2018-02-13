package themis

import (
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

func NewEtcdClient(config *EtcdConfig) {
	// create v3 client
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
}
