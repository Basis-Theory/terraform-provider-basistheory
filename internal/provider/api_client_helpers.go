package provider

func getStringPointer(value interface{}) *string {
	if str, ok := value.(string); ok {
		return &str
	}
	return nil
}

func getStringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

func getBoolPointer(value interface{}) *bool {
	if b, ok := value.(bool); ok {
		return &b
	}
	return nil
}

func getIntPointer(value interface{}) *int {
	if i, ok := value.(int); ok {
		return &i
	}
	return nil
}