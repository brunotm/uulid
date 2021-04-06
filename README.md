# UULID Universally Unique Lexicographically Sortable Identifier

![Build Status](https://github.com/brunotm/uulid/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/brunotm/uulid?cache=0)](https://goreportcard.com/report/brunotm/uulid)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/brunotm/uulid)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/brunotm/uulid/master/LICENSE)

This package provides an implementation of the [ULID spec](https://github.com/ulid/spec) compatible with the GUID/UUID format.
It is based on the amazing [oklog/ulid](https://github.com/oklog/ulid) package, with the following differences:

* Small and simple API
* Encodes to a standard UUID format and can be used as the UUID type in databases
* More safer and strict as it doesn't allow the generation time to travel backwards
* A faster monotonic RNG that don't allocate memory unnecessarily
* Defaults to text instead of binary when using the sql/driver.Valuer interface

## Install

This package requires Go modules.

```shell
go get github.com/brunotm/uulid
```

## Usage

```go
    id, err := uulid.New()
    if err != nil {
        // handle err
    }
    fmt.Println(id.String()) // Output: 0178a727-335d-db2d-4671-c5c757718d7c
```

## Test

```shell
go test ./...
```

## Command line tool

This repo also provides a tool to generate and parse UULIDs at the command line.

```shell
go get -v github.com/brunotm/ulid/cmd/uulid
```

Usage:

```shell
Usage of uulid:
  -local
        when parsing, show local time instead of UTC
  -p string
        parse the given uulid
```

## Specification

### Timestamp

* 48 bits
* UNIX-time in milliseconds
* Won't run out of space till the year 10889 AD

### Entropy

* 80 bits
* [`Monotonicity`](https://godoc.org/github.com/oklog/ulid#Monotonic) within the same millisecond

### Binary Layout and Byte Order

The components are encoded as 16 octets. Each component is encoded with the Most Significant Byte first (network byte order).

```text
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
```

### Encoding String Representation

This package encodes the ULID into the standard UUID format. **(The ULID spec uses Crockford's Base32)**

```text
UULID: 0178a73c-3cc0-71ab-74b4-cc6c7190deae

 0178a73c3cc0      71ab74b4cc6c7190deae
|------------|    |--------------------|
   Timestamp             Entropy
    12 chars             20 chars
     48bits               80bits
     base32               base32
```

## Prior Art

* [oklog/ulid](https://github.com/oklog/ulid)
* [alizain/ulid](https://github.com/alizain/ulid)
* [RobThree/NUlid](https://github.com/RobThree/NUlid)
* [imdario/go-ulid](https://github.com/imdario/go-ulid)
