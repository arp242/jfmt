package main

import (
	"fmt"
	"os"
	"strings"

	"zgo.at/jfmt"
	"zgo.at/zli"
)

var usage = `
jfmt formats JSON; https://github.com/arp242/jfmt

Usage: jfmt [-wlic] [file...]

    -w, -write      Write to file instead of stdout.
    -l, -length     Put objects and arrays on a single line if shorter than
                    this. Default is terminal width or 100. Set to 1 to always
                    use line breaks.
    -i, -indent     Indentation string; defaults to four spaces.
    -c, -color      Use colours in output: auto (default), mono, yes/always,
                    no/never.
`[1:]

func main() {
	f := zli.NewFlags(os.Args)
	var (
		help   = f.Bool(false, "h", "help")
		write  = f.Bool(false, "w", "write")
		length = f.Int(0, "l", "length")
		indent = f.String("    ", "i", "indent")
		color  = f.String("auto", "c", "color")
	)
	zli.F(f.Parse())
	if help.Bool() {
		fmt.Print(usage)
		os.Exit(0)
	}
	stdin := len(f.Args) == 0 || (len(f.Args) == 1 && f.Args[0] == "-")
	if stdin && write.Bool() {
		zli.Fatalf("cannot use -write with stdin")
	}
	if indent.String() == "\t" {
		// TODO: should probably fix this.
		zli.Fatalf("tab indents don't really work")
	}

	width := length.Int()
	if !length.Set() {
		width, _, _ = zli.TerminalSize(os.Stdout.Fd())
	}
	ff := jfmt.NewFormatter(width, "", indent.String())
	if write.Bool() {
		*color.Pointer() = "never"
	}
	switch color.String() {
	case "always", "yes", "1", "mono", "monochrome":
		zli.WantColor = true
	case "never", "no", "0":
		zli.WantColor = false
	}
	r := zli.Reset.String()
	switch color.String() {
	case "auto", "always", "yes", "1":
		ff.Highlight("key", zli.ColorHex("af5f00").String(), r)
		ff.Highlight("str", zli.ColorHex("cd0000").String(), r)
		ff.Highlight("num", zli.ColorHex("cd0000").String(), r)
		ff.Highlight("bool", zli.ColorHex("cd0000").String(), r)
		ff.Highlight("null", zli.ColorHex("008787").String(), r)
	case "mono", "monochrome":
		ff.Highlight("key", zli.Bold.String(), r)
		//ff.Highlight("str", zli.Italic.String(), r)
		//ff.Highlight("num", zli.Italic.String(), r)
		ff.Highlight("bool", zli.Underline.String(), r)
		ff.Highlight("null", zli.Underline.String(), r)
	case "never", "no", "0":
	default:
		zli.Fatalf("invalid value for -color: %q", color)
	}

	defer f.Profile()()
	if stdin {
		if zli.IsTerminal(os.Stdin.Fd()) {
			fmt.Fprintf(os.Stderr, f.Program+": reading from stdin...\r")
			os.Stderr.Sync()
		}
		zli.F(ff.Format(os.Stdout, os.Stdin))
		return
	}
	for _, file := range f.Args {
		fp, err := os.Open(file)
		zli.F(err)

		if write.Bool() {
			b := new(strings.Builder)
			zli.F(ff.Format(b, fp))
			fp.Close()
			zli.F(os.WriteFile(file, []byte(b.String()), 0o755))
		} else {
			zli.F(ff.Format(os.Stdout, fp))
			fp.Close()
		}
	}
}
