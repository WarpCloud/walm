package file

import "io/ioutil"

// MakeValueFile creates a temporary file in TempDir (see os.TempDir)
// and writes values to the file and resturn its name. It is the caller's responsibility
// to remove the file returned if necessary.
func MakeValueFile(data []byte) (string, error) {
	tmpFile, err := ioutil.TempFile("", "tmp-values-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err = tmpFile.Write(data); err != nil {
		return tmpFile.Name(), err
	}
	tmpFile.Sync()
	return tmpFile.Name(), nil
}
