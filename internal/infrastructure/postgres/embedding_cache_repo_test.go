package postgres

import "testing"

func TestVectorEncodeDecode_RoundTrip(t *testing.T) {
	in := []float32{0, 1.5, -2.25, 3.4e10}
	out, err := decodeVector(encodeVector(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(in) {
		t.Fatalf("len = %d, want %d", len(out), len(in))
	}
	for i := range in {
		if out[i] != in[i] {
			t.Errorf("out[%d] = %v, want %v", i, out[i], in[i])
		}
	}
	if _, err := decodeVector([]byte{1, 2, 3}); err == nil {
		t.Error("decodeVector of a truncated blob should fail")
	}
}
