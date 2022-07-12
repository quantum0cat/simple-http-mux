package utils

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
