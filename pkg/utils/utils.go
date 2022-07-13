package utils

import "fmt"

// RemoveDuplicates
//removes duplicates from original list and returns a new list of strings
func RemoveDuplicates(input []string) []string {
	if len(input) == 0 {
		return input
	}
	table := map[string]struct{}{}
	output := make([]string, 0, len(input))

	for _, item := range input {
		if _, exists := table[item]; !exists {
			table[item] = struct{}{}
			output = append(output, item)
		}
	}
	return output
}

//WithRid
//returns new string with request id
func WithRid(input string, rid uint32) string {
	return fmt.Sprintf("%s [rid=%X]", input, rid)
}
