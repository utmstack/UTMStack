package netblock

import "errors"

// filterSet retains every live OS ID, including failed partial construction or
// removal. A partial set is never treated as fully installed on a retry.
type filterSet struct {
	ids      []uint64
	complete bool
}

func (s filterSet) remove(remove func(uint64) error) (filterSet, error) {
	pending := s.ids[:0]
	var firstErr error
	for _, id := range s.ids {
		if err := remove(id); err != nil {
			pending = append(pending, id)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return filterSet{ids: pending}, firstErr
}

func (s filterSet) ensure(layers int, add func(int) (uint64, error), remove func(uint64) error) (filterSet, error) {
	if s.complete {
		return s, nil
	}
	s, err := s.remove(remove)
	if err != nil {
		return s, err
	}
	if layers < 1 || layers > 2 {
		return s, errors.New("invalid network filter layer count")
	}
	for i := 0; i < layers; i++ {
		id, err := add(i)
		if err != nil {
			pending, cleanupErr := s.remove(remove)
			return pending, errors.Join(err, cleanupErr)
		}
		s.ids = append(s.ids, id)
	}
	s.complete = true
	return s, nil
}
