package highlight

import (
	"bytes"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func Highlight(source, lexer, formatter, style string) (string, error) {
	// Determine lexer.
	l := lexers.Get(lexer)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := formatters.Get(formatter)
	if f == nil {
		f = formatters.Fallback
	}

	// Determine style.
	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return "", err
	}
	result := new(bytes.Buffer)
	err = f.Format(result, s, it)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
