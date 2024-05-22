package buildsql

import (
	"fmt"
	"strings"
)

// FilterBuilder struct
type FilterBuilder struct {
	prefixes []string
	filters  map[string]string
	sorts    []string
}

// NewFilterBuilder creates a new FilterBuilder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		prefixes: make([]string, 0),
		filters:  make(map[string]string),
		sorts:    []string{},
	}
}

// AddFilter adds a filter to the filter builder
func (fb *FilterBuilder) AddFilter(prefix, fieldName string, operator Operator, value string) *FilterBuilder {
	filterKey := fmt.Sprintf("%s-%s-%s", prefix, fieldName, operator)
	if fb.isValidFilter(filterKey) {
		fb.filters[filterKey] = value
		fb.prefixes = append(fb.prefixes, prefix)
	}
	return fb
}

// AddSort adds a sort to the filter builder
func (fb *FilterBuilder) AddSort(prefix, fieldName string, direction ...SortDirection) *FilterBuilder {
	if len(direction) == 0 {
		direction = append(direction, ASC)
	}

	dir := ""
	if direction[0] == DESC {
		dir = "-"
	}
	sortKey := fmt.Sprintf("%s%s-%s", dir, prefix, fieldName)
	fb.sorts = append(fb.sorts, sortKey)
	return fb
}

// isValidFilter validates the filter format
func (fb *FilterBuilder) isValidFilter(filter string) bool {
	parts := strings.Split(filter, Delimiter)
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	return len(parts) == 3
}

// String constructs the final query string
func (fb *FilterBuilder) String() string {
	var queryString strings.Builder

	// Add filters to the query string
	for field, value := range fb.filters {
		queryString.WriteString(fmt.Sprintf("filter=%s-%s&", field, value))
	}

	// Add sorts to the query string
	for _, sort := range fb.sorts {
		queryString.WriteString(fmt.Sprintf("sortOn=%s&", sort))
	}

	// Remove the trailing '&' if it exists
	result := queryString.String()
	if len(result) > 0 && result[len(result)-1] == '&' {
		result = result[:len(result)-1]
	}

	return result
}
