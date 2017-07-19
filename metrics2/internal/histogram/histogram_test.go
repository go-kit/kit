package histogram

import "testing"

func TestBucketize(t *testing.T) {
	t.Parallel()
	h := New()
	n := 1024
	PopulateNormal(h, n, 500, 25, 31337)
	for _, bucketCount := range []int{8, 16, 32} {
		buckets := h.Bucketize(bucketCount)
		if want, have := uint32(n), buckets[len(buckets)-1].Count; want != have {
			t.Errorf("Bucketize(%d): final bucket: want %d, have %d", bucketCount, want, have)
		}
	}
}

func TestRender(t *testing.T) {
	t.Parallel()
	h := New()
	PopulateNormal(h, 4096, 500, 32, 10000)
	Render(testWriter{t}, h.Bucketize(11), 40)
}

type testWriter struct{ *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.T.Logf(string(p))
	return len(p), nil
}
