package provider

import (
	"fmt"
	"strings"
)

func parseCompositeImportID(value string, expectedParts int) ([]string, error) {
	separator := "/"
	if !strings.Contains(value, separator) {
		separator = ","
	}
	parts := strings.Split(value, separator)
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("import ID has %d parts; expected %d", len(parts), expectedParts)
	}
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
		if parts[index] == "" {
			return nil, fmt.Errorf("import ID part %d is empty", index+1)
		}
	}
	return parts, nil
}
