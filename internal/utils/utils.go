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
	return []string{}
}

// GetBoolWithDefault returns the boolean value from the interface{} if it exists,
// otherwise returns the default value
func GetBoolWithDefault(value interface{}, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}

	if boolValue, ok := value.(bool); ok {
		return boolValue
	}

	return defaultValue
}

func GetOptionalParam[T any](args map[string]interface{}, key string) *T {
	if val, ok := args[key]; ok {
		if typedVal, ok := val.(T); ok {
			return &typedVal
		}
	}
	return nil
}
