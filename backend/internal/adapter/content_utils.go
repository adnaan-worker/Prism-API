package adapter

import "strings"

// GetContentAsString extracts string content from Message.Content
// Message.Content can be either string or []ContentPart (for vision)
func GetContentAsString(content interface{}) string {
	if content == nil {
		return ""
	}
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		// Handle array of content parts (vision)
		var sb strings.Builder
		for _, part := range v {
			if partMap, ok := part.(map[string]interface{}); ok {
				if text, ok := partMap["text"].(string); ok {
					sb.WriteString(text)
				}
			}
		}
		return sb.String()
	default:
		return ""
	}
}