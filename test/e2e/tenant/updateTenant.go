package node

import (
	. "github.com/onsi/ginkgo"
	"github.com/satori/go.uuid"
	"walm/pkg/tenant"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tenant", func() {

	var (
		tenantName   string
		err          error
		tenantParams tenant.TenantParams
	)

	BeforeEach(func() {

		By("create tenant")

		randomId := uuid.Must(uuid.NewV4(), err).String()
		tenantName = "test-" + randomId[:8]
		tenantParams = tenant.TenantParams{}
		err := tenant.CreateTenant(tenantName, &tenantParams)
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		err := tenant.DeleteTenant(tenantName)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("update tenant", func() {
		It("update tenant success", func() {

			tenantQuotaInfo := tenant.TenantQuotaInfo{
				LimitCpu:        "1k",
				LimitMemory:     "100Gi",
				RequestsCPU:     "1k",
				RequestsMemory:  "100Gi",
				RequestsStorage: "100Gi",
				Pods:            "1k",
			}
			tenantQuotaParam := tenant.TenantQuotaParams{QuotaName: tenantName, Hard: &tenantQuotaInfo}
			var tenantQuotaParams []*tenant.TenantQuotaParams
			tenantQuotaParams = append(tenantQuotaParams, &tenantQuotaParam)
			err := tenant.UpdateTenant(tenantName, &tenantParams)
			Expect(err).NotTo(HaveOccurred())

			// validate
			tenantInfo, err := tenant.GetTenant(tenantName)
			Expect(err).NotTo(HaveOccurred())

			for index := range tenantInfo.TenantQuotas {
				tenantQuota := tenantInfo.TenantQuotas[index]

				Expect(tenantQuota.QuotaName).To(Equal(tenantName))
				Expect(tenantQuota.Hard.LimitCpu).To(Equal("1k"))
				Expect(tenantQuota.Hard.LimitMemory).To(Equal("100Gi"))
			}

		})
	})

})
