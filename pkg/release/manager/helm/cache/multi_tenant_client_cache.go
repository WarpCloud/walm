package cache

//type MultiTenantClientsCache struct {
//	clients *lru.Cache
//}
//
//func (multiTenantClientsCache *MultiTenantClientsCache) Get(tillerHost string) *helm.Client{
//	if multiTenantClient, ok := multiTenantClientsCache.clients.Get(tillerHost); ok {
//		return multiTenantClient.(*helm.Client)
//	} else {
//		multiTenantClient = helm.NewClient(helm.Host(tillerHost))
//		multiTenantClientsCache.clients.Add(tillerHost, multiTenantClient)
//		return multiTenantClient.(*helm.Client)
//	}
//}
//
//func NewMultiTenantClientsCache(size int) *MultiTenantClientsCache {
//	multiTenantClientsCache :=  MultiTenantClientsCache{}
//	multiTenantClientsCache.clients, _ = lru.New(size)
//	return &multiTenantClientsCache
//}
