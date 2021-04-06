package uulid

import (
	"crypto/rand"
	"encoding/binary"
	"math/bits"
	"sync"
	"time"
)

// Generator implements an UUID generator based on the ULID spec.
// The generated UULID is monotonically increased for calls within the same millisecond.
type Generator struct {
	mu   sync.Mutex
	seed uint64
	ms   uint64
	hi   uint16
	lo   uint64
}

// NewGenerator is like NewGeneratorWithSeed()
// but uses a secure random seed from crypto/rand.
func NewGenerator() (r *Generator, err error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return NewGeneratorWithSeed(binary.BigEndian.Uint64(b)), nil
}

// NewGeneratorWithSeed creates a new UULID generator.
// It will used the given as the seed for the internal monotonic RNG.
//
// Ensure that a good random seed is used or use NewGenerator()
// which provides a secure seed from crypto/rand.
func NewGeneratorWithSeed(seed uint64) (r *Generator) {
	return &Generator{seed: seed}
}

// New creates a UULID with the current system time.
func (r *Generator) New() (id UULID, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ms := Timestamp(time.Now())

	if err = id.SetTimestamp(ms); err != nil {
		return id, err
	}

	if err = r.read(id[6:], ms); err != nil {
		return id, err
	}

	return id, nil
}

// read generates a pseudo random entropy that is
// incremented monotonically within the same millisecond interval
func (r *Generator) read(p []byte, ms uint64) (err error) {
	// within the same millisecond interval of the previous call
	// increment lower entropy bytes and return
	if r.ms == ms {
		lo := r.lo
		hi := r.hi

		if r.lo++; r.lo < lo {
			if r.hi++; r.hi < hi {
				return ErrMonotonicOverflow
			}
		}

		binary.BigEndian.PutUint16(p[:2], r.hi)
		binary.BigEndian.PutUint64(p[2:], r.lo)
		return nil
	}

	r.advance(ms)
	binary.BigEndian.PutUint16(p[:2], r.hi)
	binary.BigEndian.PutUint64(p[2:], r.lo)
	return nil
}

func (r *Generator) advance(ms uint64) {
	r.ms = ms
	r.hi = uint16(r.uint64r())
	r.lo = r.uint64r()
}

func (r *Generator) uint64r() (v uint64) {
	r.seed += 0xa0761d6478bd642f
	hi, lo := bits.Mul64(r.seed^0xe7037ed1a0b428db, r.seed)
	return hi ^ lo
}
