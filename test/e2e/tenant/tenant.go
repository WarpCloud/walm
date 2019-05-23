package node

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/pkg/tenant"
	"WarpCloud/walm/test/e2e/framework"
	"time"
	"WarpCloud/walm/pkg/k8s/handler"
)

var _ = Describe("Tenant", func() {

	var (
		tenantName string
	)

	BeforeEach(func() {
		tenantName = framework.GenerateRandomName("tenantTest")
	})

	AfterEach(func() {
		framework.DeleteNamespace(tenantName, false)
	})

	It("test tenant lifecycle: create, update, delete", func() {

		tenantQuotaInfo := tenant.TenantQuotaInfo{
			LimitCpu:        "10",
			LimitMemory:     "50Gi",
			RequestsCPU:     "10",
			RequestsMemory:  "50Gi",
			RequestsStorage: "100Gi",
			Pods:            "100",
		}
		tenantQuotaParams := []*tenant.TenantQuotaParams{{QuotaName: "walm-default", Hard: &tenantQuotaInfo}}
		tenantParams := &tenant.TenantParams{TenantQuotas: tenantQuotaParams}
		err := tenant.CreateTenant(tenantName, tenantParams)
		Expect(err).NotTo(HaveOccurred())

		tenantInfo, err := getTenant(tenantName)
		Expect(err).NotTo(HaveOccurred())
		Expect(*tenantInfo.TenantQuotas[0].Hard).To(Equal(tenantQuotaInfo))

		_, err = handler.GetDefaultHandlerSet().GetLimitRangeHandler().GetLimitRangeFromK8s(tenantName, tenant.LimitRangeDefaultName)
		Expect(err).NotTo(HaveOccurred())

		tenantQuotaInfo.LimitCpu = "20"
		err = tenant.UpdateTenant(tenantName, tenantParams)
		Expect(err).NotTo(HaveOccurred())
		tenantInfo, err = getTenant(tenantName)
		Expect(err).NotTo(HaveOccurred())
		Expect(*tenantInfo.TenantQuotas[0].Hard).To(Equal(tenantQuotaInfo))

		err = tenant.DeleteTenant(tenantName)
		Expect(err).NotTo(HaveOccurred())
	})

})

func getTenant(tenantName string) (*tenant.TenantInfo, error){
	time.Sleep(time.Second * 1)
	return tenant.GetTenant(tenantName)
}
