//go:build linux && amd64

package config

import (
	"maps"
)

func updateOptional[T comparable](dst **T, newVal *T) bool {
	if newVal == nil {
		if *dst == nil {
			return false
		}
		*dst = nil // Clear the destination pointer
		return true
	}
	val := *newVal // Dereference newVal to get the value
	if *dst != nil && **dst == val {
		return false // If values are the same, no update needed
	}
	// Allocate a fresh copy so callers can't mutate our internal state
	valCopy := val
	*dst = &valCopy // Store pointer to the new copy
	return true
}

func deepCopyMap[K comparable, V any](oldMap map[K]V) map[K]V {
	if oldMap == nil {
		return nil
	}
	newMap := make(map[K]V, len(oldMap))
	maps.Copy(newMap, oldMap)
	return newMap
}
