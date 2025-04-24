package utils

func GetStringSlice(arg map[string]interface{}, key string) []string {
	if val, ok := arg[key]; ok && val != nil {
		if slice, ok := val.([]string); ok {
			return slice
		}
		// fallback: it might be []interface{} with string values
		if ifaceSlice, ok := val.([]interface{}); ok {
			strSlice := make([]string, 0, len(ifaceSlice))
			for _, v := range ifaceSlice {
				if str, ok := v.(string); ok {
					strSlice = append(strSlice, str)
				}
			}
			return strSlice
		}
	}
	return nil
}
