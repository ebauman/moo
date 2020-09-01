package rulestore

import (
	"github.com/ebauman/moo/pkg/types"
	"sort"
)

type Store struct {
	rules []types.Rule
}

func NewStore() *Store {
	return &Store{
		rules: make([]types.Rule, 0),
	}
}

func (s *Store) AddRule(r types.Rule) bool {
	s.rules = append(s.rules, types.Rule{})
	if len(s.rules) == 0 {
		s.rules = append(s.rules, r)
		return true
	}

	i := sort.Search(len(s.rules), func(i int) bool {
		return s.rules[i].Priority < r.Priority
	})

	copy(s.rules[i+1:], s.rules[i:])

	s.rules[i] = r

	return true
}

func (s *Store) ListRules() []types.Rule {
	return s.rules
}

func (s *Store) DeleteRule(index int) bool {
	s.rules = append(s.rules[:index], s.rules[index+1:]...)

	return true
}