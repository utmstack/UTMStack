package scanner

import "context"

// Tests inject the same generation-bearing API as production. Capture identity
// before the modeled analysis, so a reload inside it cannot relabel the result.
func snapshotForTest(s *Scanner, scan func(context.Context, []byte) (bool, string, error)) func(context.Context, []byte) (bool, string, string, error) {
	return func(ctx context.Context, data []byte) (bool, string, string, error) {
		generation := s.sigDBVersion()
		clean, signature, err := scan(ctx, data)
		return clean, signature, generation, err
	}
}
