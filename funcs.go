package sq

import (
	"errors"
	"fmt"
	"net/url"
	pathpkg "path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	douceur "github.com/aymerick/douceur/parser"
	"github.com/emptyinterface/ago"
	otto "github.com/robertkrimen/otto/parser"
)

type (
	ParseFunc func(s, arg string) (string, error)

	LoadFunc func(sel *goquery.Selection, s, arg string) (interface{}, error)

	TypeLoader struct {
		isType func(t reflect.Type) bool
		load   func(sel *goquery.Selection, s string) (interface{}, error)
	}

	parser struct {
		f    ParseFunc
		args string
	}
	loader struct {
		f    LoadFunc
		args string
	}
)

var (
	ErrNoRegexpMatch = errors.New("regexp did not match the content")
)

var (
	parseFuncs = map[string]ParseFunc{
		"regexp": func(s, pattern string) (string, error) {
			r, err := regexp.Compile(pattern)
			if err != nil {
				return "", err
			}
			matches := r.FindStringSubmatch(s)
			if len(matches) == 1 {
				return matches[0], nil
			}
			if len(matches) > 1 {
				if len(matches[1]) == 0 {
					return "", ErrNoRegexpMatch
				}
				return matches[1], nil
			}
			return "", ErrNoRegexpMatch
		},
		"strip": func(s, pattern string) (string, error) {
			r, err := regexp.Compile(pattern)
			if err != nil {
				return "", err
			}
			return r.ReplaceAllString(s, ""), nil
		},
		"path.prepend": func(s, token string) (string, error) {
			if strings.HasPrefix(s, token) {
				return s, nil
			}
			fmt.Println(pathpkg.Join(token, s))
			return pathpkg.Join(token, s), nil
		},
		"path.append": func(s, token string) (string, error) {
			if strings.HasSuffix(s, token) {
				return s, nil
			}
			fmt.Println(pathpkg.Join(s, token))
			return pathpkg.Join(s, token), nil
		},
		"prepend": func(s, token string) (string, error) {
			if strings.HasPrefix(s, token) {
				return s, nil
			}
			return token + s, nil
		},
		"append": func(s, token string) (string, error) {
			if strings.HasSuffix(s, token) {
				return s, nil
			}
			return s + token, nil
		},
	}

	loadFuncs = map[string]LoadFunc{
		"time": func(_ *goquery.Selection, s, layout string) (interface{}, error) {
			return time.Parse(strings.TrimSpace(layout), strings.TrimSpace(s))
		},
		"ago": func(_ *goquery.Selection, s, _ string) (interface{}, error) {
			return ago.Parse(strings.TrimSpace(s))
		},
	}

	typeLoaders = map[string]TypeLoader{
		"url": {
			isType: func(t reflect.Type) bool {
				return t.PkgPath() == "net/url" && t.Name() == "URL"
			},
			load: func(_ *goquery.Selection, s string) (interface{}, error) {
				return url.Parse(s)
			},
		},
		"goquery": {
			isType: func(t reflect.Type) bool {
				return strings.HasSuffix(t.PkgPath(), "/goquery") && t.Name() == "Selection"
			},
			load: func(sel *goquery.Selection, _ string) (interface{}, error) {
				return sel.Clone(), nil
			},
		},
		"html": {
			isType: func(t reflect.Type) bool {
				return t.PkgPath() == "golang.org/x/net/html" && t.Name() == "Node"
			},
			load: func(sel *goquery.Selection, _ string) (interface{}, error) {
				return sel.Clone().Nodes[0], nil
			},
		},
		"otto": {
			isType: func(t reflect.Type) bool {
				return strings.HasSuffix(t.PkgPath(), "otto/ast") && t.Name() == "Program"
			},
			load: func(sel *goquery.Selection, text string) (interface{}, error) {
				if text == "" {
					text = sel.Text()
				}
				return otto.ParseFile(nil, "", text, 0)
			},
		},
		"css": {
			isType: func(t reflect.Type) bool {
				return strings.HasSuffix(t.PkgPath(), "douceur/css") && t.Name() == "Stylesheet"
			},
			load: func(sel *goquery.Selection, text string) (interface{}, error) {
				if text == "" {
					text = sel.Text()
				}
				return douceur.Parse(text)
			},
		},
	}
)

func RegisterParseFunc(name string, f ParseFunc) {
	parseFuncs[name] = f
}

func RegisterLoadFunc(name string, f LoadFunc) {
	loadFuncs[name] = f
}

func RegisterTypeLoader(name string, isType func(t reflect.Type) bool, load func(sel *goquery.Selection, text string) (interface{}, error)) {
	typeLoaders[name] = TypeLoader{
		isType: isType,
		load:   load,
	}
}

func (p parser) parse(s string) (string, error) {
	if p.f != nil {
		return p.f(s, p.args)
	}
	return s, nil
}

func (l *loader) load(sel *goquery.Selection, s string) (interface{}, error) {
	if l.f != nil {
		return l.f(sel, s, l.args)
	}
	return s, nil
}
