package sq

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractString(t *testing.T) {

	const (
		title    = "test title"
		attr     = "test attr"
		fragment = "<a>test fragment</a>"
	)

	var testHTML = fmt.Sprintf(`
		<html>
			<head><title>%s</title></head>
			<body><p data-attr=%q>%s</p></body>
		</html>
	`, title, attr, fragment)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Error(err)
	}

	s, err := extractString(doc.Find("title"), accessorText)
	if err != nil {
		t.Error(err)
	}
	if s != title {
		t.Errorf("Expected %q, got %q", title, s)
	}

	s, err = extractString(doc.Find("[data-attr]"), fmt.Sprintf("%s(data-attr)", accessorAttr))
	if err != nil {
		t.Error(err)
	}
	if s != attr {
		t.Errorf("Expected %q, got %q", attr, s)
	}

	expected := fmt.Errorf("%s: %v", ErrAttributeNotFound, "attr(missing)")
	s, err = extractString(doc.Find("[data-attr]"), fmt.Sprintf("%s(missing)", accessorAttr))
	if err.Error() != expected.Error() {
		t.Errorf("Expected %q, got %q", expected, err)
	}
	if s != "" {
		t.Errorf("Expected empty string, got %q", s)
	}

	s, err = extractString(doc.Find("p"), accessorHTML)
	if err != nil {
		t.Error(err)
	}
	if s != fragment {
		t.Errorf("Expected %q, got %q", fragment, s)
	}

	expected = fmt.Errorf("Bad accessor: %q", "madeupaccessor")
	s, err = extractString(doc.Find("p"), "madeupaccessor")
	if err.Error() != expected.Error() {
		t.Errorf("Expected %q, got %q", expected, err)
	}
	if s != "" {
		t.Errorf("Expected empty string, got %q", s)
	}

}

func TestParseTag(t *testing.T) {

	tests := []struct {
		tag reflect.StructTag
		p   *path
		err error
	}{
		// good
		{`sq:"p.last"`, &path{selector: "p.last"}, nil},
		{`sq:"p.last | text"`, &path{selector: "p.last", acc: "text"}, nil},
		{`sq:"p.last | text | regexp(\\d+)"`, &path{
			selector: "p.last", acc: "text",
			parsers: []parser{parser{args: "\\d+", f: parseFuncs["regexp"]}}},
			nil,
		},
		{`sq:"p.last | text | regexp(\\d+) | regexp(\\d)"`,
			&path{
				selector: "p.last", acc: "text",
				parsers: []parser{
					parser{args: "\\d+", f: parseFuncs["regexp"]},
					parser{args: "\\d", f: parseFuncs["regexp"]},
				},
			},
			nil,
		},
		{`sq:"p.last | text | regexp(\\d+) | regexp(\\d) | time(01)"`,
			&path{
				selector: "p.last", acc: "text",
				parsers: []parser{
					parser{args: "\\d+", f: parseFuncs["regexp"]},
					parser{args: "\\d", f: parseFuncs["regexp"]},
				},
				loader: &loader{args: "01", f: loadFuncs["time"]},
			},
			nil,
		},
		{`sq:"p.last | attr(id) | regexp(\\d+) | regexp(\\d) | time(01)"`,
			&path{
				selector: "p.last", acc: "attr(id)",
				parsers: []parser{
					parser{args: "\\d+", f: parseFuncs["regexp"]},
					parser{args: "\\d", f: parseFuncs["regexp"]},
				},
				loader: &loader{args: "01", f: loadFuncs["time"]},
			},
			nil,
		},
		{`sq:"() p"`, &path{flags: []string{}, selector: "p"}, nil},
		{`sq:"( ) p"`, &path{flags: []string{}, selector: "p"}, nil},
		{`sq:"(opt1) p"`, &path{flags: []string{"opt1"}, selector: "p"}, nil},
		{`sq:"( opt1 ) p"`, &path{flags: []string{"opt1"}, selector: "p"}, nil},
		{`sq:"(opt1,opt2) p"`, &path{flags: []string{"opt1", "opt2"}, selector: "p"}, nil},
		{`sq:"( opt1 , opt2 ) p"`, &path{flags: []string{"opt1", "opt2"}, selector: "p"}, nil},
		{`sq:"(opt1,opt2,opt3) p"`, &path{flags: []string{"opt1", "opt2", "opt3"}, selector: "p"}, nil},
		{`sq:"(    opt1 ,opt2, opt3    )    p"`, &path{flags: []string{"opt1", "opt2", "opt3"}, selector: "p"}, nil},

		// bad
		{`sq:"p.last | fuzzy"`, nil, fmt.Errorf("Bad accessor: %q", `fuzzy`)},
		{`sq:"p.last | text | unregifunc"`, nil, fmt.Errorf("%q not registered func", "unregifunc")},
		{`sq:"p.last\d"`, nil, fmt.Errorf("Bad tag: %q", `sq:"p.last\d"`)},
		{``, nil, ErrTagNotFound},
	}

	for _, test := range tests {
		p, err := parseTag(test.tag)
		if err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("Expected error %q, got %q", test.err, err)
			}
			continue
		}
		if len(p.flags) != len(test.p.flags) {
			t.Errorf("Expected %#v, got %#v", test.p.flags, p.flags)
		} else {
			for i := 0; i < len(p.flags); i++ {
				if p.flags[i] != test.p.flags[i] {
					t.Errorf("Expected %#v, got %#v", test.p.flags, p.flags)
				}
			}
		}
		if p.selector != test.p.selector {
			t.Errorf("Expected %q, got %q", test.p.selector, p.selector)
		}
		if p.acc != test.p.acc {
			t.Errorf("Expected %q, got %q", test.p.acc, p.acc)
		}
		if p.loader == nil && p.loader != test.p.loader {
			t.Errorf("Expected %#v, got %#v", test.p.loader, p.loader)
		}
		if p.loader != nil {
			if test.p.loader.args != p.loader.args {
				t.Errorf("Expected %#v, got %#v", test.p.loader.args, p.loader.args)
			}
			if reflect.ValueOf(test.p.loader.f).Pointer() != reflect.ValueOf(p.loader.f).Pointer() {
				t.Errorf("Expected %#v, got %#v", test.p.loader, p.loader)
			}
		}
		if len(p.parsers) != len(test.p.parsers) {
			t.Errorf("Expected %#v, got %#v", test.p.parsers, p.parsers)
		}
		for i, pp := range p.parsers {
			if test.p.parsers[i].args != pp.args {
				t.Errorf("Expected %q, got %q", test.p.parsers[i].args, pp.args)
			}
			if reflect.ValueOf(test.p.parsers[i].f).Pointer() != reflect.ValueOf(pp.f).Pointer() {
				t.Errorf("Expected %#v, got %#v", test.p.parsers[i].f, pp.f)
			}
		}
	}

}
