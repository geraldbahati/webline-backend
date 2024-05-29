package utils

import (
	"fmt"
	"strings"
)

// Filter struct to manage the filters
type Filter struct {
	filters map[string][]string
	args    []interface{}
}

// NewFilter initializes a new Filter
func NewFilter() *Filter {
	return &Filter{
		filters: make(map[string][]string),
		args:    []interface{}{},
	}
}

// HasFilter checks if there are any filters
func (f *Filter) HasFilter() bool {
	return len(f.filters) > 0
}

// Add adds a new filter to the Filter
func (f *Filter) Add(key, operator string, offset int, value interface{}) {
	filter := fmt.Sprintf("%s %s $%d", key, operator, len(f.args)+(offset+1))
	f.filters[key] = append(f.filters[key], filter)
	f.args = append(f.args, value)
}

// AddRaw adds a raw filter string
func (f *Filter) AddRaw(key, rawFilter string) {
	f.filters[key] = []string{rawFilter}
}

// GetParameterizedQuery returns the filters as a SQL WHERE clause and arguments
func (f *Filter) GetParameterizedQuery() (string, []interface{}) {
	parts := []string{}
	for _, filter := range f.filters {
		groupedValues := strings.Join(filter, " OR ")
		parts = append(parts, fmt.Sprintf("(%s)", groupedValues))
	}
	return strings.Join(parts, " AND "), f.args
}
