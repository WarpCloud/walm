package redis

const (
	WalmReleasesKey   = "walm-releases"
	WalmProjectsKey   = "walm-project-tasks"
	WalmReleaseTasksKey   = "walm-release-tasks"
)

type Redis interface {
	GetFieldValue(key, namespace, name string) (string, error)
	GetFieldValues(key, namespace string) ([]string, error)
	GetFieldValuesByNames(key string, filedNames... string) ([]string, error)
	SetFieldValues(key string, fieldValues map[string]interface{}) error
	DeleteField(key, namespace, name string) error
}

func BuildFieldName(namespace, name string) string {
	return namespace + "/" + name
}