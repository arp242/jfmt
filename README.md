jfmt is my JSON formatter. There's a few things this does different than most
JSON formatters I could find, which in my opinion makes the output much nicer.

Install with `go install zgo.at/jfmt/cmd/jfmt@latest`.

Or import `zgo.at/jfmt` to use in Go code.

---

It tries to be a little smarter about what should be on a single line. Things
like this:

    {
        "nums": [
            12,
            4,
            901,
            12,
            742,
        ],
        "strings": [
            "k",
            "p",
            "bas",
            "pqz",
            "q21",
        ],
        "obj" {
            "k": "v"
        }
    }

Are just silly, and splitting it out over many lines like that isn't helpful. So that
becomes:

    {
        "nums":    [12, 4, 901, 12, 742],
        "strings": ["k", "p", "bas", "pqz", "q21"],
        "obj"      {"k": "v"}
    }

It will still write to multiple lines if it's too long.

It will also align the keys; which I rather like.

Like many JSON formatting tools keys are sorted alphabetically, but it puts
values that fit on a single line first, to avoid avoid things like this:

    {
      "additional": false,
      "definitions": {
          [.. very long object ..]
       }
      "title": "JSON schema for Python project metadata and configuration",
      "type": "object",
    }

I don't like these kind of "dangling primitives", so that that becomes:

    {
      "additional": false,
      "title":      "JSON schema for Python project metadata and configuration",
      "type":       "object",
      "definitions": {
          [.. very long object ..]
       }
    }

And things like this:

    {
        "arr5": [
            [
                [
                    [
                        [
                            {"type": "string", "value": "#"}
                        ]
                    ]
                ]
            ]
        ]
    }

Can just be written on a single line, too:

    {
        "arr5": [[[[[{"type": "string", "value": "#"}]]]]],
    }

There's a bunch of cases like this; in general it tries to be as concise as
possible as long as it doesn't sacrifice readability.

---

The downside of this is that it's a bit slower and uses more memory. It's still
plenty fast enough for almost all use cases; the largest file in schemastore.org
(kubernetes-definitions.json, 17,500 lines/950K) takes about 120ms on my laptop
(vs. 30ms with Go's default MarshalIndent). I spent almost zero effort
optimising anything, so there's probably a lot to win here.
