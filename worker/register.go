package worker

import (
	"context"
	"github.com/k-si/crongo/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"net"
	"time"
)

var Register WorkRegistry

type WorkRegistry struct {
	IP string
}

func RegistryWorker(ctx context.Context) (err error) {
	var ip string
	if ip, err = getLocalIP(); err != nil {
		return
	}
	Register = WorkRegistry{
		IP: ip,
	}
	go Register.Registry(ctx)
	return
}

func (r WorkRegistry) Registry(ctx context.Context) {

	var (
		err           error
		grantResp     *clientv3.LeaseGrantResponse
		keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
	)

	wkey := common.WorkerDir + r.IP

	for {
		cancelCtx, cancelFunc := context.WithCancel(context.Background())
		if grantResp, err = EtcdConn.cli.Grant(context.TODO(), 5); err != nil {
			goto retry
		}
		if keepAliveChan, err = EtcdConn.cli.KeepAlive(cancelCtx, grantResp.ID); err != nil {
			goto retry
		}
		if _, err = EtcdConn.cli.Put(context.TODO(), wkey, ""); err != nil {
			goto retry
		}
		for {
			select {
			case <-ctx.Done():
				cancelFunc()
				goto end
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil {
					goto retry
				}
			}
		}
	retry:
		log.Println("register lease can not keep alive, start retry after 5 second")
		cancelFunc()
		time.Sleep(5 * time.Second)
	}
end:
}

func getLocalIP() (ip string, err error) {
	var (
		addrs []net.Addr
		ipnet *net.IPNet
		ok    bool
	)

	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	for _, addr := range addrs {
		if ipnet, ok = addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip = ipnet.IP.To4().String()
		}
	}
	if ip == "" {
		err = common.ErrNotFoundIP
	}

	return
}
