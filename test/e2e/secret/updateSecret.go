package secret

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/satori/go.uuid"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
	"os"
	"go/build"
	"encoding/base64"
)

var _ = Describe("Secret", func() {

	var (
		gopath    string
		namespace string
		randomId  string
		secretName string

	)

	BeforeEach(func() {

		By("create namespace")

		gopath = os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}

		randomId = uuid.Must(uuid.NewV4()).String()
		namespace = "test-" + randomId[:8]
		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      namespace,
			},
		}
		_, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&ns)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("update secret", func() {

		It("update secret success", func() {

			// create a secret
			secretName = "token-" + randomId[:5]

			walmSecret := &adaptor.WalmSecret{
				Type: "Opaque",
				WalmMeta: adaptor.WalmMeta{
					Namespace: namespace,
					Name: secretName,
				},
			}

			err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor).CreateSecret(walmSecret)
			Expect(err).NotTo(HaveOccurred())

			// data is a map of strings to []byte, which serialize as base64 (even in json)
			data := map[string]string{}
			data["namespace"] = base64.StdEncoding.EncodeToString([]byte(namespace))

			walmSecret = &adaptor.WalmSecret{
				Type: "Opaque",
				Data: data,
				WalmMeta: adaptor.WalmMeta{
					Namespace: namespace,
					Name: secretName,
				},
			}

			// update a secret
			err = adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor).UpdateSecret(walmSecret)
			Expect(err).NotTo(HaveOccurred())

			// validate updated data info
			secrets, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor).ListSecrets(namespace, nil)
			for index := range secrets.Items {
				if secretName == secrets.Items[index].Name {
					newData := secrets.Items[index].Data
					Expect(newData["namespace"]).To(Equal(namespace))
				}
			}

		})
	})
})
