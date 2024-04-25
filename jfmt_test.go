package jfmt

import (
	"os"
	"strings"
	"testing"

	"zgo.at/zstd/ztest"
)

// TODO: toml-test-key-escapes and toml-test-key-space have the wrong alignmnet
//
// TODO: toml-test-inline-table-nest is kinda meh
func Test(t *testing.T) {
	ls, err := os.ReadDir("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range ls {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			in, err := os.ReadFile("testdata/" + f.Name())
			if err != nil {
				t.Fatal(err)
			}

			have, err := NewFormatter(100, "    ").FormatString(string(in))
			if err != nil {
				t.Fatal(err)
			}
			if d := ztest.Diff(have, string(in)); d != "" {
				t.Fatal(d)
			}
		})
	}
}

func Benchmark(b *testing.B) {
	d, err := os.ReadFile("testdata/schemastore-global.json")
	if err != nil {
		b.Fatal(err)
	}

	ff := NewFormatter(80, "    ")
	in := string(d)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ff.FormatString(in)
	}
}
