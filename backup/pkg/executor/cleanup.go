package executor

import (
	"sort"
)

func cleanup(limit int, files []string) []string {
	sortedFiles := files
	sort.Strings(sortedFiles)

	if len(sortedFiles) > limit {
		return sortedFiles[:len(sortedFiles)-limit]
	}

	return []string{}
}
