package util


func MergeValues(dest map[string]interface{}, src map[string]interface{}, deleteKey bool) map[string]interface{} {
	for k, v := range src {
		if deleteKey && v == nil{
			delete(dest, k)
			continue
		}

		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = MergeValues(destMap, nextMap, deleteKey)
	}
	return dest
}
