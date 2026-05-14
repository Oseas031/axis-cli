package tmp

import "testing"

func TestStack(t *testing.T) {
	s := &Stack{}
	if s.Len() != 0 { t.Errorf("empty stack len = %d", s.Len()) }
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if s.Len() != 3 { t.Errorf("len after 3 pushes = %d", s.Len()) }
	if v := s.Pop(); v != 3 { t.Errorf("pop = %d, want 3", v) }
	if v := s.Pop(); v != 2 { t.Errorf("pop = %d, want 2", v) }
	if s.Len() != 1 { t.Errorf("len after 2 pops = %d", s.Len()) }
	s.Pop()
	if v := s.Pop(); v != -1 { t.Errorf("pop empty = %d, want -1", v) }
}
