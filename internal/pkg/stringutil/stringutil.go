// Package stringutil implements string utilities.
package stringutil

import (
	"sort"
	"strings"
	"unicode"
)

// TrimLines splits the output into individual lines and trims the spaces from each line.
//
// This also trims the start and end spaces from the original output.
func TrimLines(output string) string {
	split := strings.Split(strings.TrimSpace(output), "\n")
	lines := make([]string, 0, len(split))
	for _, line := range split {
		lines = append(lines, strings.TrimSpace(line))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// MapToSortedSlice transforms m to a sorted slice.
func MapToSortedSlice(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))
	for e := range m {
		s = append(s, e)
	}
	sort.Strings(s)
	return s
}

// SliceToMap transforms s to a map.
func SliceToMap(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, e := range s {
		m[e] = struct{}{}
	}
	return m
}

// SliceToUniqueSortedSlice returns a sorted copy of s with no duplicates.
func SliceToUniqueSortedSlice(s []string) []string {
	return MapToSortedSlice(SliceToMap(s))
}

// SliceToUniqueSortedSliceFilterEmptyStrings returns a sorted copy of s with no duplicates and no empty strings.
//
// Strings with only spaces are considered empty.
func SliceToUniqueSortedSliceFilterEmptyStrings(s []string) []string {
	m := SliceToMap(s)
	for key := range m {
		if strings.TrimSpace(key) == "" {
			delete(m, key)
		}
	}
	return MapToSortedSlice(m)
}

// SliceToChunks splits s into chunks of the given chunk size.
//
// If s is nil or empty, returns empty.
// If chunkSize is <=0, returns [][]string{s}.
func SliceToChunks(s []string, chunkSize int) [][]string {
	var chunks [][]string
	if len(s) == 0 {
		return chunks
	}
	if chunkSize <= 0 {
		return [][]string{s}
	}
	c := make([]string, len(s))
	copy(c, s)
	// https://github.com/golang/go/wiki/SliceTricks#batching-with-minimal-allocation
	for chunkSize < len(c) {
		c, chunks = c[chunkSize:], append(chunks, c[0:chunkSize:chunkSize])
	}
	return append(chunks, c)
}

// SnakeCaseOption is an option for snake_case conversions.
type SnakeCaseOption func(*snakeCaseOptions)

// SnakeCaseWithNewWordOnDigits is a SnakeCaseOption that signfies
// to split on digits, ie foo_bar_1 instead of foo_bar1.
func SnakeCaseWithNewWordOnDigits() SnakeCaseOption {
	return func(snakeCaseOptions *snakeCaseOptions) {
		snakeCaseOptions.newWordOnDigits = true
	}
}

// ToLowerSnakeCase transforms s to lower_snake_case.
func ToLowerSnakeCase(s string, options ...SnakeCaseOption) string {
	return strings.ToLower(toSnakeCase(s, options...))
}

// ToUpperSnakeCase transforms s to UPPER_SNAKE_CASE.
func ToUpperSnakeCase(s string, options ...SnakeCaseOption) string {
	return strings.ToUpper(toSnakeCase(s, options...))
}

// ToPascalCase converts s to PascalCase.
//
// Splits on '-', '_', ' ', '\t', '\n', '\r'.
// Uppercase letters will stay uppercase,
func ToPascalCase(s string) string {
	output := ""
	var previous rune
	for i, c := range strings.TrimSpace(s) {
		if !isDelimiter(c) {
			if i == 0 || isDelimiter(previous) || unicode.IsUpper(c) {
				output += string(unicode.ToUpper(c))
			} else {
				output += string(unicode.ToLower(c))
			}
		}
		previous = c
	}
	return output
}

func toSnakeCase(s string, options ...SnakeCaseOption) string {
	snakeCaseOptions := &snakeCaseOptions{}
	for _, option := range options {
		option(snakeCaseOptions)
	}
	output := ""
	s = strings.TrimFunc(s, isDelimiter)
	for i, c := range s {
		if isDelimiter(c) {
			c = '_'
		}
		if i == 0 {
			output += string(c)
		} else if isSnakeCaseNewWord(c, snakeCaseOptions.newWordOnDigits) &&
			output[len(output)-1] != '_' &&
			((i < len(s)-1 && !isSnakeCaseNewWord(rune(s[i+1]), true) && !isDelimiter(rune(s[i+1]))) ||
				(snakeCaseOptions.newWordOnDigits && unicode.IsDigit(c)) ||
				(unicode.IsLower(rune(s[i-1])))) {
			output += "_" + string(c)
		} else if !(isDelimiter(c) && output[len(output)-1] == '_') {
			output += string(c)
		}
	}
	return output
}

func isSnakeCaseNewWord(r rune, newWordOnDigits bool) bool {
	if newWordOnDigits {
		return unicode.IsUpper(r) || unicode.IsDigit(r)
	}
	return unicode.IsUpper(r)
}

func isDelimiter(r rune) bool {
	return r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

type snakeCaseOptions struct {
	newWordOnDigits bool
}
