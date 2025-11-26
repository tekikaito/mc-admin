package utils

import (
	"strings"
)

func SplitAndTrim(commaSeperatedElements string, seperator string) []string {
	var element []string
	for _, name := range strings.Split(commaSeperatedElements, seperator) {
		element = append(element, strings.TrimSpace(name))
	}
	return element
}
