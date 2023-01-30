package utils

import (
	"context"
	"github.com/grydovee/ingress-frp/pkg/constants"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
	"time"
)

type DomainWatcher struct {
	resolver *net.Resolver

	addr           string
	ips            []net.IP
	OnClientChange func([]net.IP)
}

func NewDomainWatcher(addr string) *DomainWatcher {
	return &DomainWatcher{
		resolver: &net.Resolver{},
		addr:     addr,
	}
}

func (w *DomainWatcher) Start(ctx context.Context) {
	w.syncClients(ctx)
	ticker := time.NewTicker(constants.DomainSyncInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.syncClients(ctx)
		}
	}
}

func (w *DomainWatcher) syncClients(ctx context.Context) {
	ipAddrs, err := w.resolver.LookupIPAddr(ctx, w.addr)
	if err != nil {
		log.FromContext(ctx).Error(err, "lookup addr error")
		return
	}
	ips := make([]net.IP, len(ipAddrs))
	for i, addr := range ipAddrs {
		ips[i] = addr.IP
	}
	sort.Slice(ips, func(i, j int) bool {
		// todo compare bits
		return ips[i].String() < ips[j].String()
	})
	if ipsEqual(ips, w.ips) {
		return
	}
	w.ips = ips
	w.OnClientChange(ips)
}

func ipsEqual(ips1, ips2 []net.IP) bool {
	if len(ips1) != len(ips2) {
		return false
	}
	for i, ip := range ips1 {
		if !ip.Equal(ips2[i]) {
			return false
		}
	}
	return true
}
