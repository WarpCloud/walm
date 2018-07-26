package transwarp

import (
	"fmt"
	jsonnet "github.com/google/go-jsonnet"
	"path/filepath"
)

type MemoryImporter struct {
	Data map[string]string
}

func (importer *MemoryImporter) Import(dir, importedPath string) (*jsonnet.ImportedData, error) {
	path := filepath.Join(dir, importedPath)
	if content, ok := importer.Data[path]; ok {
		return &jsonnet.ImportedData{Content: content, FoundHere: path}, nil
	}
	return nil, fmt.Errorf("Import not available %v", path)
}

func MakeMemoryVM(data map[string]string) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.Importer(&MemoryImporter{Data: data})
	return vm
}
