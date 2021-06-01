package uulid_test

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/brunotm/uulid"
)

var (
	// Time: 2021-04-05T14:51:43.663UTC Timestamp: 1617634303663
	encoded   = []byte("0178a284-9eaf-b3e7-036d-5b1b9f3cd753")
	timestamp = uint64(1617634303663)
	entropy   = "b3e7036d5b1b9f3cd753"
)

func TestNew_UULID(t *testing.T) {
	_, err := uulid.New()
	if err != nil {
		t.Error(err)
	}

}

func TestUULID_SetTimestamp(t *testing.T) {
	id, err := uulid.New()
	if err != nil {
		t.Error(err)
	}

	if err = id.SetTimestamp(timestamp); err != nil {
		t.Error(err)
	}

	if err = id.SetTime(time.Now()); err != nil {
		t.Error(err)
	}

	if err = id.SetTimestamp(uulid.MaxTimestamp + 1); err != uulid.ErrBigTime {
		t.Errorf("expected ErrBigTime, got %s instead", err)
	}
}

func TestUULID_String(t *testing.T) {
	id, err := uulid.Parse(encoded)
	if err != nil {
		t.Error(err)
	}

	if id.String() != string(encoded) {
		t.Errorf("parse error, expected: %s, got: %s", string(encoded), id.String())
	}

}

func TestUULID_Entropy(t *testing.T) {
	id, err := uulid.Parse(encoded)
	if err != nil {
		t.Error(err)
	}

	if hex.EncodeToString(id.Entropy()) != entropy {
		t.Errorf("parse error, expected: %s, got: %s", entropy, id.String())
	}

}

func TestUULID_Compare(t *testing.T) {
	id1, err := uulid.New()
	if err != nil {
		t.Error(err)
	}

	id2, err := uulid.New()
	if err != nil {
		t.Error(err)
	}

	if id1.Compare(id2) != -1 {
		t.Errorf("compare error, expected: %d, got: %d", -1, id1.Compare(id2))
	}

	if id2.Compare(id1) != 1 {
		t.Errorf("compare error, expected: %d, got: %d", 1, id1.Compare(id2))
	}

	if id1.Compare(id1) != 0 {
		t.Errorf("compare error, expected: %d, got: %d", -1, id1.Compare(id1))
	}

}

func TestUULID_Marshaler(t *testing.T) {
	id, err := uulid.Parse(encoded)
	if err != nil {
		t.Error(err)
	}

	if v, err := id.Value(); err != nil || v.(string) != string(encoded) {
		t.Error(err)
	}

	if x, err := id.MarshalBinary(); err != nil || !bytes.Equal(x, id[:]) {
		t.Errorf("not equal: %s and %s, error: %s", x, string(id[:]), err)
	}

	if x, err := id.MarshalText(); err != nil || !bytes.Equal(x, encoded) {
		t.Errorf("not equal: %s and %s, error: %s", x, encoded, err)
	}

	buf := make([]byte, 6)
	if err = id.MarshalBinaryTo(buf); err != uulid.ErrBufferSize {
		t.Errorf("expected ErrBufferSize, got: %s", err)
	}

	if err = id.MarshalTextTo(buf); err != uulid.ErrBufferSize {
		t.Errorf("expected ErrBufferSize, got: %s", err)
	}

}

func TestUULID_Unmarshaler(t *testing.T) {
	id, err := uulid.Parse(encoded)
	if err != nil {
		t.Error(err)
	}
	id2 := id

	if err = id.Scan(encoded); err != nil || !bytes.Equal(id[:], id2[:]) {
		t.Errorf("not equal: %s and %s, error: %s", id.String(), string(encoded), err)
	}

	var buf []byte
	if buf, err = id.MarshalBinary(); err != nil {
		t.Error(err)
	}

	if err = id.UnmarshalBinary(buf); err != nil {
		t.Error(err)
	}

	if err = id.UnmarshalText(encoded); err != nil {
		t.Error(err)
	}

	if err = id.UnmarshalBinary(encoded); err != uulid.ErrDataSize {
		t.Error(err)
	}

	if err = id.UnmarshalText([]byte(`123456789090`)); err != uulid.ErrDataSize {
		t.Error(err)
	}

}

func TestParse(t *testing.T) {
	id, err := uulid.Parse(encoded)
	if err != nil {
		t.Error(err)
	}

	if id.Compare(uulid.UULID{}) == 0 {
		t.Error("non-initialized uulid")
	}

	if id.Timestamp() != timestamp {
		t.Errorf("time parse error, expected: %d, got: %d", timestamp, id.Timestamp())
	}

	if hex.EncodeToString(id.Entropy()) != entropy {
		t.Errorf("time parse error, expected: %s, got: %s", entropy, hex.EncodeToString(id.Entropy()))
	}
}

func TestTimestamp(t *testing.T) {
	tm := uulid.Time(timestamp)
	ts := uulid.Timestamp(tm)

	if ts != timestamp {
		t.Errorf("time parse error, expected: %d, got: %d", timestamp, ts)
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(uulid.BinarySize)

	for i := 0; i < b.N; i++ {
		_, _ = uulid.New()
	}
}

func BenchmarkParse(b *testing.B) {
	id, err := uulid.New()
	if err != nil {
		b.Fatal(err)
	}

	buf := []byte(id.String())

	b.ReportAllocs()
	b.SetBytes(uulid.HexEncodedSize)

	for i := 0; i < b.N; i++ {
		_, _ = uulid.Parse(buf)
	}
}

func BenchmarkNewConcurrent(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(uulid.BinarySize)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = uulid.New()
		}
	})
}
