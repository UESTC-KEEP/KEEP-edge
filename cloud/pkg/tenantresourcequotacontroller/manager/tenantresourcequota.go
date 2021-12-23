package manager

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"keep/cloud/pkg/tenantresourcequotacontroller/config"
	"sync"
)

// EqualNodeManager 监测EqualNode crd事件的manager
type TenantResourceQuotaManager struct {
	// events 从apiserver收到的事件
	events chan watch.Event
	// EqualNode 键值对 equalnode.nodename:*types.EqualNode
	EqualNode sync.Map
}

// Events 获取该类型的所有事件
func (eqndmm *TenantResourceQuotaManager) Events() chan watch.Event {
	return eqndmm.events
}

// NewEqualNodeManager 注册一个controller监听事件的增删改
func NewEqualNodeManager(si cache.SharedInformer) (*TenantResourceQuotaManager, error) {
	events := make(chan watch.Event, config.Config.Buffer.EqualNodeEvent)
	rh := NewCommonResourceEventHandler(events)
	si.AddEventHandler(rh)
	return &TenantResourceQuotaManager{events: events}, nil
}
