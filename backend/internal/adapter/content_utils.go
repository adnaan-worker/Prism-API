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

// parseDataURL parses a data URL and returns media type and base64 data
// Format: data:image/png;base64,xxxxx
func parseDataURL(url string) (mediaType, data string) {
	// Check if it's a data URL
	if !strings.HasPrefix(url, "data:") {
		return "", ""
	}

	// Remove "data:" prefix
	url = strings.TrimPrefix(url, "data:")

	// Find the comma separating media type from data
	commaIdx := strings.Index(url, ",")
	if commaIdx == -1 {
		return "", ""
	}

	// Extract media type and data
	mediaTypeAndEncoding := url[:commaIdx]
	data = url[commaIdx+1:]

	// Remove encoding suffix (e.g., ";base64")
	if semiIdx := strings.Index(mediaTypeAndEncoding, ";"); semiIdx != -1 {
		mediaType = mediaTypeAndEncoding[:semiIdx]
	} else {
		mediaType = mediaTypeAndEncoding
	}

	return mediaType, data
}