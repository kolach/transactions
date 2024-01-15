package db

import (
	"fmt"
)

// stringToInt32Ptr converts a string to an int32 pointer.
func stringToInt32Ptr(s string) (*int32, error) {
	if s == "" {
		return nil, nil
	}

	var i int32
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to int32: %w", err)
	}

	return &i, nil
}
