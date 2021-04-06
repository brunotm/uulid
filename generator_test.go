package uulid_test

import (
	"testing"

	"github.com/brunotm/uulid"
)

func TestNew_Generator(t *testing.T) {
	var err error
	if _, err = uulid.NewGenerator(); err != nil {
		t.Error(err)
	}
}

func TestGenerator_New(t *testing.T) {
	r, err := uulid.NewGenerator()
	if err != nil {
		t.Error(err)
	}

	id, err := r.New()
	if err != nil {
		t.Error(err)
	}

	if id.Compare(uulid.UULID{}) == 0 {
		t.Error("non-initialized uulid")
	}
}

func BenchmarkTestGenerator_SeqSafety(b *testing.B) {
	r, err := uulid.NewGenerator()
	if err != nil {
		b.Error(err)
	}

	b.ReportAllocs()
	b.SetBytes(uulid.BinarySize)

	b.RunParallel(func(pb *testing.PB) {
		prev, err := r.New()
		if err != nil {
			b.Error(err)
		}

		for pb.Next() {
			cur, err := r.New()
			if err != nil {
				b.Error(err)
			}

			if prev.Compare(cur) != -1 {
				b.Error(prev.Compare(cur), prev.Time(), cur.Time())
			}

			prev = cur
		}

	})

}
