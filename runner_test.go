package testdown_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	. "github.com/jmalloc/testdown"
)

func TestRunner(t *testing.T) {
	l := &Loader{
		FS: os.DirFS("."),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	test, err := l.Load(ctx, "testdata")
	if err != nil {
		t.Fatal(err)
	}

	r := &NativeRunner{
		Output: func(a Assertion) (string, error) {
			input := []byte(a.Input)

			var v any
			if err := json.Unmarshal(input, &v); err != nil {
				return "", err
			}

			output, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return "", err
			}

			return string(output) + "\n", nil
		},
	}

	r.Run(t, test)
}
