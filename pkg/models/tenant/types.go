package tenant

const (
	MultiTenantLabelKey = "multi-tenant"
)

type TenantInfoList struct {
	Items []*TenantInfo `json:"items" description:"tenant list"`
}

//Tenant Info
type TenantInfo struct {
	TenantName            string                  `json:"tenantName" description:"name of the tenant"`
	TenantCreationTime    string                  `json:"tenantCreationTime" description:"create time of the tenant"`
	TenantLabels          map[string]string       `json:"tenantLabels"  description:"labels of the tenant"`
	TenantAnnotitions     map[string]string       `json:"tenantAnnotations"  description:"annotations of the tenant"`
	TenantStatus          string                  `json:"tenantStatus" description:"status of the tenant"`
	TenantQuotas          []*TenantQuota          `json:"tenantQuotas" description:"quotas of the tenant"`
	MultiTenant           bool                    `json:"multiTenant" description:"multi tenant"`
	Ready                 bool                    `json:"ready" description:"tenant ready status"`
	UnifyUnitTenantQuotas []*UnifyUnitTenantQuota `json:"unifyUnitTenantQuotas" description:"quotas of the tenant with unified unit"`
}

type UnifyUnitTenantQuota struct {
	QuotaName string                    `json:"quotaName" description:"quota name"`
	Hard      *UnifyUnitTenantQuotaInfo `json:"hard" description:"quota hard limit"`
	Used      *UnifyUnitTenantQuotaInfo `json:"used" description:"quota used"`
}

//Tenant Params Info
type TenantParams struct {
	TenantAnnotations map[string]string    `json:"tenantAnnotations"  description:"annotations of the tenant"`
	TenantLabels      map[string]string    `json:"tenantLabels"  description:"labels of the tenant"`
	TenantQuotas      []*TenantQuotaParams `json:"tenantQuotas" description:"quotas of the tenant"`
}

type TenantQuotaParams struct {
	QuotaName string           `json:"quotaName" description:"quota name"`
	Hard      *TenantQuotaInfo `json:"hard" description:"quota hard limit"`
}

type TenantQuota struct {
	QuotaName string           `json:"quotaName" description:"quota name"`
	Hard      *TenantQuotaInfo `json:"hard" description:"quota hard limit"`
	Used      *TenantQuotaInfo `json:"used" description:"quota used"`
}

//Quota Info
type TenantQuotaInfo struct {
	LimitCpu        string `json:"limitCpu"  description:"requests of the CPU"`
	LimitMemory     string `json:"limitMemory"  description:"limit of the memory"`
	RequestsCPU     string `json:"requestsCpu"  description:"requests of the CPU"`
	RequestsMemory  string `json:"requestsMemory"  description:"requests of the memory"`
	RequestsStorage string `json:"requestsStorage"  description:"requests of the storage"`
	Pods            string `json:"pods" description:"num of the pods"`
}

type UnifyUnitTenantQuotaInfo struct {
	LimitCpu        float64 `json:"limitCpu"  description:"requests of the CPU"`
	LimitMemory     int64   `json:"limitMemory"  description:"limit of the memory"`
	RequestsCPU     float64 `json:"requestsCpu"  description:"requests of the CPU"`
	RequestsMemory  int64   `json:"requestsMemory"  description:"requests of the memory"`
	RequestsStorage int64   `json:"requestsStorage"  description:"requests of the storage"`
	Pods            int64   `json:"pods" description:"num of the pods"`
}
