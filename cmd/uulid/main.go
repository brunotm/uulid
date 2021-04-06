package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/brunotm/uulid"
)

const (
	rfc3339ms = "2006-01-02T15:04:05.999MST"
)

var (
	p     = flag.String("p", "", "parse the given uulid")
	local = flag.Bool("local", false, "when parsing, show local time instead of UTC")
)

func main() {
	flag.Parse()

	switch *p {
	case "":
		id, err := uulid.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%s\n", id.String())

	default:
		id, err := uulid.Parse([]byte(*p))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		t := id.Time()
		if !*local {
			t = t.UTC()
		}

		fmt.Fprintf(os.Stderr, "Time: %s,  Timestamp: %d, Entropy: %s\n",
			t.Format(rfc3339ms),
			id.Timestamp(),
			hex.EncodeToString(id.Entropy()))
	}
}
