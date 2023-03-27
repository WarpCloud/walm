package main

import "github.com/pkg/errors"

var errNamespaceRequired = errors.New("flag --namespace/-n required")
var errServerRequired = errors.New("flag --server/-s required")
var errProjectNameRequired = errors.New("flag --name required for specify projectName")

func checkResourceType(sourceType string) error {
	if sourceType != "release" && sourceType != "project" && sourceType != "migration" && sourceType != "namespace"{
		return errors.Errorf("the server doesn't have a resource type %s, release, project, migration only", sourceType)
	}
	return nil
}
