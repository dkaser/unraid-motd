package utils

// StringSet a set for strings, useful for keeping track of elements
type StringSet map[string]struct{}

// Contains returns true if `v` is in the set
func (s StringSet) Contains(v string) bool {
	_, ok := s[v]

	return ok
}

// FromList returns a `StringSet` with the input list's contents
func (s StringSet) FromList(listIn []string) StringSet {
	var empty struct{}
	p := make(StringSet)
	for _, val := range listIn {
		p[val] = empty
	}

	return p
}
