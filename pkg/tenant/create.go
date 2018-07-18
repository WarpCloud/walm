package tenant

import (
	"errors"

	. "walm/pkg/util/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateTenant initialize the namespace for the tenant
// and installs the essential components
func CreateTenant(name, namespace string) error {
	err := createTiller(namespace, "3", true, false, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func createTiller(namespace, historyMax string, rbac, security bool, labels, annotations map[string]string) error {
	// TODO: delete this if statement after security is implemented
	if security == true {
		return errors.New("security is not implemented yet")
	}
	if rbac == true {
		serviceAccount := corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "tiller",
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
			},
		}
		role := v1beta1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: "rbac.authorization.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "tiller",
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
			},
			Rules: []v1beta1.PolicyRule{
				{
					Verbs:     []string{"*"},
					APIGroups: []string{"*"},
					Resources: []string{"*"},
				},
			},
		}
		roleBinding := v1beta1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "tiller",
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
			},
			Subjects: []v1beta1.Subject{
				{
					Kind:      "ServiceAccount",
					APIGroup:  "rbac.authorization.k8s.io",
					Name:      "tiller",
					Namespace: namespace,
				},
			},
			RoleRef: v1beta1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "tiller",
			},
		}

		//add by bing han 2018/7/16 to pass build
		Log.Infof("%v,%v,%v", serviceAccount, role, roleBinding)
		//add end

		// TODO: create these resources
	}
	if security == true {
		// TODO: implement security here
	}
	// TODO: add install tiller via helm here
	// helm init --service-account tiller --ca...
	return nil
}
