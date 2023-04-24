package util

type Set map[string]bool

func (s Set) Add(key string) {
	s[key] = true
}

func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Del(key string) {
	delete(s, key)
}

func (left Set) Equals(right Set) bool {
	if len(left) != len(right) {
		return false
	}

	for key := range left {
		if !right.Has(key) {
			return false
		}
	}

	return true
}

func (left Set) EqualsStrs(right []string) bool {
	return left.Equals(NewSet(right...))
}

func NewSet(elements ...string) Set {
	set := Set{}

	for _, element := range elements {
		set.Add(element)
	}

	return set
}
