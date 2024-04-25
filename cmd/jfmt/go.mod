// Make this a separate module so importing jfmt doesn't pull in zli, x/term,
// and x/sys.

module zgo.at/jfmt/cmd/jfmt

go 1.22

require (
	zgo.at/jfmt v0.0.0-00010101000000-000000000000
	zgo.at/zli v0.0.0-20240425054714-1cba1e6760ff
)

replace zgo.at/jfmt => ../../

require (
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/term v0.14.0 // indirect
	zgo.at/termtext v1.3.0 // indirect
	zgo.at/zstd v0.0.0-20240425000522-78bcf900e0a4 // indirect
)
