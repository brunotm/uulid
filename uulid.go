// Package uulid provides an UUID compatible ULID
package uulid

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"time"
)

const (
	MaxTimestamp   = 281474976710655
	HexEncodedSize = 36
	BinarySize     = 16
)

var (

	// ErrBigTime is returned if the given time that is larger than MaxTimestamp.
	ErrBigTime = errors.New("uulid: time greater than supported by the ulid spec")

	// ErrDataSize is returned when parsing an invalid string representation of a UULID.
	ErrDataSize = errors.New("uulid: bad data size when parsing")

	// ErrBufferSize is returned when marshalling an UULID to a buffer < 36 bytes.
	ErrBufferSize = errors.New("uulid: bad buffer size when marshaling")

	// ErrInvalidType is returned when scan receives an invalid type.
	ErrInvalidType = errors.New("uulid: invalid type to unmarshal")

	// ErrMonotonicOverflow is returned if the current 10bit entropy overflows.
	ErrMonotonicOverflow = errors.New("uulid: monotonic overflow")

	// ErrSmallTime is returned if the current epoch time is lower than the previously seen
	// by the Generator.
	ErrSmallTime = errors.New("uulid: time is lower than current generator")

	// generator is the default Generator for the package
	generator *Generator
)

func init() {
	var err error
	if generator, err = NewGenerator(); err != nil {
		panic(err)
	}
}

/*
An UULID is a 16 byte Universally Unique Lexicographically Sortable Identifier

	The components are encoded as 16 octets.
	Each component is encoded with the MSB first (network byte order).

	0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                      32_bit_uint_time_high                    |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|     16_bit_uint_time_low      |       16_bit_uint_random      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                       32_bit_uint_random                      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                       32_bit_uint_random                      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type UULID [BinarySize]byte

func New() (id UULID, err error) {
	return generator.New()
}

// Time returns the UULID time component with a millisecond precision
func (id UULID) Time() time.Time {
	return Time(id.Timestamp())
}

// Timestamp return the UULID millisecond unix timestamp
func (id UULID) Timestamp() uint64 {
	// Adapted from binary.BigEndian.Uint64 to 6 byte
	return uint64(id[5]) |
		uint64(id[4])<<8 |
		uint64(id[3])<<16 |
		uint64(id[2])<<24 |
		uint64(id[1])<<32 |
		uint64(id[0])<<40

}

// SetTimestamp sets the time component of the ULID to the given Unix time
// in milliseconds.
func (id *UULID) SetTimestamp(ms uint64) (err error) {
	if ms > MaxTimestamp {
		return ErrBigTime
	}

	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)

	return nil
}

// SetTime sets the time component of the ULID to the given time.Time.
func (id *UULID) SetTime(t time.Time) (err error) {
	return id.SetTimestamp(Timestamp(t))
}

// SetEntropy sets the ULID entropy to the passed byte slice.
func (id *UULID) SetEntropy(e []byte) (err error) {
	if len(e) != 10 {
		return ErrDataSize
	}

	copy(id[6:], e)
	return nil
}

// Compare returns an integer comparing id and other lexicographically.
// The result will be 0 if id==other, -1 if id < other, and +1 if id > other.
func (id UULID) Compare(other UULID) (i int) {
	return bytes.Compare(id[:], other[:])
}

// String returns the string encoded UULID
func (id *UULID) String() (s string) {
	b := make([]byte, HexEncodedSize)
	id.MarshalTextTo(b)
	return string(b)
}

// MarshalBinaryTo writes the binary encoding of the ULID to the given buffer.
// ErrBufferSize is returned when the len(dst) != 16.
func (id UULID) MarshalBinaryTo(dst []byte) (err error) {
	if len(dst) != BinarySize {
		return ErrBufferSize
	}

	copy(dst, id[:])
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (id UULID) MarshalBinary() (data []byte, err error) {
	data = make([]byte, BinarySize)
	return data, id.MarshalBinaryTo(data)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (id *UULID) UnmarshalBinary(data []byte) (err error) {
	if len(data) != BinarySize {
		return ErrDataSize
	}

	return parse(data, id)
}

// MarshalTextTo writes the UULID as a string to the given buffer.
// ErrBufferSize is returned when the len(dst) != EncodedSize.
func (id UULID) MarshalTextTo(dst []byte) (err error) {
	if len(dst) != HexEncodedSize {
		return ErrBufferSize
	}

	hex.Encode(dst[0:8], id[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], id[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], id[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], id[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], id[10:16])
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (id UULID) MarshalText() (data []byte, err error) {
	data = make([]byte, HexEncodedSize)
	return data, id.MarshalTextTo(data)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (id *UULID) UnmarshalText(data []byte) (err error) {
	return parse(data, id)
}

// MarshalJSONTo writes the UULID as a string to the given buffer.
// ErrBufferSize is returned when len(dst) != EncodedSize.
func (id UULID) MarshalJSONTo(dst []byte) (err error) {
	return id.MarshalTextTo(dst)
}

// MarshalJSON implements the json.Marshaler interface.
func (id UULID) MarshalJSON() (data []byte, err error) {
	return id.MarshalText()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (id *UULID) UnmarshalJSON(data []byte) (err error) {
	return parse(data, id)
}

// Scan implements the sql.Scanner interface.
// It supports scanning a string or byte slice.
func (id *UULID) Scan(src interface{}) (err error) {
	switch x := src.(type) {
	case nil:
		return nil
	case string:
		return parse([]byte(x), id)
	case []byte:
		return parse(x, id)
	}

	return ErrInvalidType
}

// Value implements the sql/driver.Valuer
func (id UULID) Value() (v driver.Value, err error) {
	b, err := id.MarshalText()
	return string(b), err
}

// Entropy returns the entropy from the UULID.
func (id UULID) Entropy() (data []byte) {
	data = make([]byte, 10)
	copy(data, id[6:])
	return data
}

// Parse parses an encoded UULID, returning an error in case of failure.
//
// ErrDataSize is returned if the length is different from an encoded
// UULID valid lengths, either 32 or 36 characters.
//
// ErrBigTime is returned if time is greater than MaxTime().
func Parse(data []byte) (id UULID, err error) {
	err = parse(data, &id)
	return id, err
}

func parse(data []byte, id *UULID) (err error) {
	switch len(data) {
	case 16: // binary encoded
		copy(id[:], data)

	case 32: // UUID hex format 0177de6a6f3dd1d5f5f7d0c250314de9
		if _, err = hex.Decode(id[:], data); err != nil {
			return err
		}

	case 36: // UUID standard format 0177de6a-6f3d-d1d5-f5f7-d0c250314de9
		if _, err := hex.Decode(id[0:4], data[0:8]); err != nil {
			return err
		}
		if _, err := hex.Decode(id[4:6], data[9:13]); err != nil {
			return err
		}
		if _, err := hex.Decode(id[6:8], data[14:18]); err != nil {
			return err
		}
		if _, err := hex.Decode(id[8:10], data[19:23]); err != nil {
			return err
		}
		if _, err := hex.Decode(id[10:16], data[24:36]); err != nil {
			return err
		}

	default:
		return ErrDataSize
	}

	if id.Timestamp() > MaxTimestamp {
		return ErrBigTime
	}

	return nil
}

// Timestamp converts a time.Time to Unix milliseconds.
// Times from the year 10889 produces undefined results.
func Timestamp(t time.Time) (ms uint64) {
	return uint64(t.Unix())*1000 +
		uint64(t.Nanosecond()/int(time.Millisecond))
}

// Time converts Unix milliseconds in the format
// returned by the Timestamp function to a time.Time.
func Time(ms uint64) (t time.Time) {
	return time.Unix(int64(ms/1e3), int64((ms%1e3)*1e6))
}

// MaxTime returns the maximum time supported by an UULID
func MaxTime() (t time.Time) { return Time(MaxTimestamp) }
