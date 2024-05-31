package jfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"zgo.at/termtext"
	"zgo.at/zstd/zbyte"
	"zgo.at/zstd/zmap"
)

type Formatter struct {
	indentStr           string
	width               int
	hlKey, hlStr, hlNum [2]string
	hlBool, hlNull      [2]string

	// TODO: don't have mutable state; pass in function instead.
	didTop        bool
	keyCol, level int
	key           string
}

func NewFormatter(width int, ident string) *Formatter {
	return &Formatter{width: width, indentStr: ident}
}

var widthCache = make(map[string]int)

func init() {
	termtext.Widths = map[rune]int{
		'\b': 2,
		'\f': 2,
		'\n': 2,
		'\r': 2,
		'\t': 2,
	}
}

// Calling termtext.Width is pretty slow; about ~75% of all time is spent
// there. Caching the results improves things a bit.
//
// TODO: be a bit smarter so we just don't call termtext.Width() that often.
func strWidth(s string) int {
	if w, ok := widthCache[s]; ok {
		return w
	}
	w := termtext.Width(s)
	widthCache[s] = w
	return w
}

// Highlight sets the syntax highlighting.
//
// what can be key, str, num, bool, or null.
//
// Default is to not highlight anything.
func (f *Formatter) Highlight(what, start, stop string) {
	switch what {
	case "key":
		f.hlKey = [2]string{start, stop}
	case "str":
		f.hlStr = [2]string{start, stop}
	case "num":
		f.hlNum = [2]string{start, stop}
	case "bool":
		f.hlBool = [2]string{start, stop}
	case "null":
		f.hlNull = [2]string{start, stop}
	default:
		panic(fmt.Sprintf("jfmt.Highlight: unknown: %q", what))
	}
}

// Format the reader to the writer.
func (f *Formatter) Format(w io.Writer, r io.Reader) error {
	f.didTop = false
	if f.width == 0 {
		f.width = 100
	}
	in, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	{
		// TODO: this is a hack to allow "JSON lines"; should probably just include
		// a copy/fork of json package.
		//
		// Also, would like better error reporting; another reason to fork.
		pos := zbyte.IndexAll(in, []byte("}\n{"))
		if len(pos) > 0 {
			prev := 0
			pos = append(pos, len(in)-2)
			for _, p := range pos {
				jj := in[prev : p+1]
				prev = p + 2

				var j any
				err = json.Unmarshal(jj, &j)
				if err != nil {
					return err
				}
				f.any(w, j)
				fmt.Fprint(w, "\n")
				f.didTop = false
			}
			return nil
		}
	}

	var j any
	err = json.Unmarshal(in, &j)
	if err != nil {
		return err
	}
	// TODO: this should probably return write errors to w, but meh
	f.any(w, j)
	fmt.Fprint(w, "\n")
	return nil
}

// Format the string.
func (f *Formatter) FormatString(j string) (string, error) {
	b := new(strings.Builder)
	err := f.Format(b, strings.NewReader(j))
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (f *Formatter) indent(mod int) string {
	f.level += mod
	return strings.Repeat(f.indentStr, f.level)

}

func (f *Formatter) any(w io.Writer, j any) bool {
	switch jj := j.(type) {
	default:
		panic(fmt.Sprintf("jfmt: unknown type: %T", jj)) /// Should never happen.
	case nil:
		fmt.Fprint(w, f.hlNull[0], "null", f.hlNull[1])
		return false
	case bool:
		if jj {
			fmt.Fprint(w, f.hlBool[0], "true", f.hlBool[1])
		} else {
			fmt.Fprint(w, f.hlBool[0], "false", f.hlBool[1])
		}
		return false
	case float64:
		fl := strconv.FormatFloat(jj, 'f', 0, 64)
		fmt.Fprint(w, f.hlNum[0], fl, f.hlNum[1])
		return false
	case string:
		s, _ := json.Marshal(jj)
		fmt.Fprint(w, `"`, f.hlStr[0], string(s[1:len(s)-1]), f.hlStr[1], `"`)
		return false
	case map[string]any:
		if !f.didTop {
			f.didTop = true
			if len(jj) == 0 { /// special case for top-level empty objects
				fmt.Fprint(w, "{}")
				return false
			}
			return f.objNL(w, jj)
		}
		return f.obj(w, jj)
	case []any:
		if !f.didTop {
			f.didTop = true
			return f.arr(w, jj)
		}
		return f.arr(w, jj)
	}
}

func (f *Formatter) arr(w io.Writer, a []any) bool {
	var hasobj, multiline bool
	if len(a) > 1 {
		for _, aa := range a {
			if _, ok := aa.(map[string]any); ok {
				hasobj, multiline = true, true
				break
			}
			if _, ok := aa.([]any); ok {
				hasobj, multiline = true, true
				break
			}
		}
	}

	var (
		start = f.keyCol + strWidth(f.key) + len(f.indentStr)*f.level +
			5 /// Quotes, :, space, [
		l = start
	)
	if start < 0 {
		panic(fmt.Sprintf("NEGATIVE START (%d): f.keyCol=%d  f.key=%q  f.indentStr=%q  f.level=%d",
			start, f.keyCol, f.key, f.indentStr, f.level))
	}
	fmt.Fprint(w, "[")
	if hasobj {
		fmt.Fprint(w, "\n", f.indent(+1))
	}
	for i, aa := range a {
		if i > 0 {
			if hasobj {
				fmt.Fprint(w, ",\n", f.indent(0))
			} else if l > 0 && l > f.width {
				multiline = true
				fmt.Fprint(w, ",\n", strings.Repeat(" ", start))
				l = start
			} else {
				fmt.Fprint(w, ", ")
			}
		}
		b := new(bytes.Buffer)
		f.any(b, aa)
		l += strWidth(b.String()) + 2
		w.Write(b.Bytes())
	}
	if hasobj {
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, f.indent(-1))
	}
	fmt.Fprint(w, "]")
	return multiline
}

func (f *Formatter) obj(w io.Writer, m map[string]any) bool {
	/// Don't nest multiple objects on the same line.
	for _, v := range m {
		if _, ok := v.(map[string]any); ok {
			return f.objNL(w, m)
		}
	}

	var (
		l = len(f.indentStr) * f.level
		b = new(strings.Builder)
	)
	fmt.Fprint(b, "{")
	for i, k := range zmap.KeysOrdered(m) {
		if i > 0 {
			fmt.Fprint(b, ", ")
		}
		kk, _ := json.Marshal(k)
		fmt.Fprint(b, `"`, f.hlKey[0], string(kk[1:len(kk)-1]), f.hlKey[1], `": `)
		f.any(b, m[k])

		if strWidth(b.String())+l > f.width {
			return f.objNL(w, m)
		}
	}
	fmt.Fprint(b, "}")
	w.Write([]byte(b.String()))
	return false
}

func (f *Formatter) objNL(w io.Writer, m map[string]any) bool {
	// We need to process all the keys since we want to only sort *multiline*
	// collections last.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	var (
		output = make(map[string]string, len(m))
		multi  = make(map[string]int8, len(m))
		l      int
	)
	fmt.Fprint(w, "{\n")
	f.level++
	kc := l
	f.keyCol += l
	for _, k := range keys {
		f.key = k
		b := new(strings.Builder)
		if f.any(b, m[k]) {
			multi[k] = 1
		} else if ll := strWidth(k); ll > l {
			l = ll /// Don't align multiline values.
		}
		output[k] = b.String()
	}
	l += 2 /// Quote marks and :

	// TODO: maybe also sort on object size (longer ones go later)? Need
	// to see how well that works.
	sort.Strings(keys)
	sort.SliceStable(keys, func(i, j int) bool { return multi[keys[i]] < multi[keys[j]] })
	// Go 1.21
	//slices.Sort(keys)
	//slices.SortStableFunc(keys, func(a, b string) int { return cmp.Compare(multi[a], multi[b]) })

	for i, k := range keys {
		if i > 0 {
			fmt.Fprint(w, ",\n")
		}
		kk, _ := json.Marshal(k)
		var pad string
		// TODO: this should also pad multiline arrays, but only if it contains
		// wrapped values (as in schemastore-global.json) and not if every entry
		// is on its one line (as in toml-test-spec-example-1.json)
		//
		// The trick bit here is correctly aligning the indent; we don't have
		// this information yet when writing the value above.
		if ll := l - strWidth(string(kk)); multi[k] == 0 && ll > 0 {
			pad = strings.Repeat(" ", ll)
		}
		fmt.Fprint(w, f.indent(0),
			`"`, f.hlKey[0], string(kk[1:len(kk)-1]), f.hlKey[1], `": `, pad,
			output[k])
	}

	f.keyCol -= kc
	fmt.Fprint(w, "\n", f.indent(-1), "}")
	return true
}
