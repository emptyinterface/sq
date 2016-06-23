package sq

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type (
	path struct {
		selector string
		acc      string
		parsers  []parser
		loader   *loader
	}
)

const (
	accessorAttr = "attr"
	accessorHTML = "html"
	accessorText = "text"
)

func extractString(sel *goquery.Selection, acc string) (string, error) {

	switch {
	case acc == accessorHTML:
		s, err := sel.Html()
		return strings.TrimSpace(s), err
	case acc == accessorText:
		return strings.TrimSpace(sel.Text()), nil
	case strings.HasPrefix(acc, accessorAttr):
		s, exists := sel.Attr(trimAccessor(acc, accessorAttr))
		if !exists {
			return "", fmt.Errorf("%s: %v", ErrAttributeNotFound, acc)
		}
		return strings.TrimSpace(s), nil
	// jank
	case acc == "":
		return "", nil
	}

	return "", fmt.Errorf("Bad accessor: %q", acc)

}

func trimAccessor(s, prefix string) string {
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(s, ")")
	return strings.TrimSpace(s)
}

func parseFunctionSignature(s string) (string, string) {

	var name, args string

	if i := strings.IndexByte(s, '('); i > -1 {
		name = s[:i]
		if j := strings.LastIndexByte(s, ')'); j > -1 && i < j {
			args = s[i+1 : j]
		}
	} else {
		name = s
	}

	return name, args

}

func parseTag(tag reflect.StructTag) (*path, error) {
	p := &path{}
	for i, part := range strings.Split(tag.Get("sq"), " | ") {
		part = strings.TrimSpace(part)
		switch i {
		case 0:
			p.selector = part
		case 1:
			switch {
			case strings.HasPrefix(part, accessorAttr+"("),
				accessorHTML == part,
				accessorText == part:
				p.acc = part
			default:
				return nil, fmt.Errorf("Bad accessor: %q", part)
			}
		default:
			name, args := parseFunctionSignature(part)
			if pf, exists := parseFuncs[name]; exists {
				p.parsers = append(p.parsers, parser{f: pf, args: args})
			} else if lf, exists := loadFuncs[name]; exists {
				p.loader = &loader{lf, args}
			} else {
				return nil, fmt.Errorf("%q not registered func", name)
			}
		}
	}
	if p.selector == "" {
		if strings.Contains(string(tag), "sq:") {
			return nil, fmt.Errorf("Bad tag: %q", tag)
		}
		return nil, ErrTagNotFound
	}
	return p, nil
}
