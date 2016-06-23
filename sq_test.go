package sq

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/emptyinterface/sq/test"
	"github.com/robertkrimen/otto/ast"
)

type errreader struct{}

var errreadererr = errors.New("errreader")

func (_ errreader) Read(_ []byte) (int, error) { return 0, errreadererr }

func TestFails(t *testing.T) {

	var notapointer struct{}
	errs := Scrape(notapointer, strings.NewReader(""))
	if len(errs) != 1 || errs[0] != ErrNonStructPtrValue {
		t.Errorf("Expected %q, got %q", ErrNonStructPtrValue, errs)
	}

	errs = Scrape(&notapointer, errreader{})
	if len(errs) != 1 || errs[0] != errreadererr {
		t.Errorf("Expected %q, got %q", errreadererr, errs)
	}

}

func TestText(t *testing.T) {

	const testHTML = `
		<html>
			<head>
				<script type="text/javascript">var a = 1;</script>
				<script type="text/javascript">var a = 2;</script>
				<script type="text/javascript">var a = 3;</script>
				<style>a { font-size: 1px; }</style>
				<style>a { font-size: 2px; }</style>
				<style>a { font-size: 3px; }</style>
			</head>
			<body>
			<table class="list">
				<tr><td>1</td><td>2</td><td>3</td></tr>
				<tr><td>11</td><td>22</td><td>33</td></tr>
				<tr><td>111</td><td>222</td><td>333</td></tr>
			</table>
			<div>
				<a href="https://www.google.com">∆</a>
				<p>0</p>
				<p>1</p>
				<p>2</p>
			</div>
			<p class="array">0</p>
			<p class="array">1</p>
			<p class="array">2</p>
			<p class="array">3</p>
			<p class="slice">0.1</p>
			<p class="slice">1.1</p>
			<p class="slice">2.1f</p>
			<p class="byteslice">ByteSlice</p>
			<p class="eightbytearray">EightByteArray</p>
			<p class="bool">true</p>
			<p class="byte">8</p>
			<p class="int">-48</p>
			<p class="int8">8</p>
			<p class="int16">16</p>
			<p class="int32">32</p>
			<p class="int64">64</p>
			<p class="uint">48</p>
			<p class="uint8">8</p>
			<p class="uint16">16</p>
			<p class="uint32">32</p>
			<p class="uint64">64</p>
			<p class="uintptr">255</p>
			<p class="float32">1.234</p>
			<p class="float64">2.468</p>
			<p class="interface">Interface</p>
			<p class="string">String</p>
			<p class="time">The date today is: 2016 05 23</p>
		</body></html>
	`
	RegisterParseFunc("nestedfail", func(_, _ string) (string, error) {
		return "", errors.New("nested fail")
	})
	RegisterParseFunc("parsefail", func(_, _ string) (string, error) {
		return "", errors.New("parse fail")
	})
	RegisterLoadFunc("loadfail", func(_ *goquery.Selection, _, _ string) (interface{}, error) {
		return "", errors.New("load fail")
	})

	RegisterTypeLoader("customtype",
		func(t reflect.Type) bool {
			return strings.HasSuffix(t.PkgPath(), "test") && t.Name() == "CustomType"
		},
		func(_ *goquery.Selection, text string) (interface{}, error) {
			return test.CustomType(text), nil
		},
	)

	var expectederrs = []string{
		`invalid kind: map`,
		`p.int: strconv.ParseBool: parsing "-48": invalid syntax`,
		`p.bool: strconv.ParseInt: parsing "true": invalid syntax`,
		`p.bool: strconv.ParseUint: parsing "true": invalid syntax`,
		`p.bool: strconv.ParseFloat: parsing "true": invalid syntax`,
		`p.bool: (loader fail) "parsing time \"true\": extra text: true"`,
		`attribute not found: attr(missing)`,
		`attribute not found: attr(missing)`,
		`attribute not found: attr(missing)`,
		`Bad tag: "sq:\"derp(\\d)\""`,
		`p.bool: (parser fail) "parse fail"`,
		`p.bool: (loader fail) "load fail"`,
		`div: (parser fail) "nested fail"`,
		`div: (parser fail) "nested fail"`,
		`private field with sq tag: "privatetagged"`,
		`"blink" did not match`,
		`"blink.selection" did not match`,
		`"blink.node" did not match`,
		`"blink.javascript" did not match`,
		`"blink.css" did not match`,
		`Bad accessor: "badacc.goquery"`,
		`Bad accessor: "badacc.node"`,
		`Bad accessor: "badacc.url"`,
		`Bad accessor: "badacc.javascript"`,
		`Bad accessor: "badacc.css"`,
		`a: (parser fail) "parse fail"`,
		`a: (parser fail) "parse fail"`,
		`a: (parser fail) "parse fail"`,
		`a: (parser fail) "parse fail"`,
		`a: (parser fail) "parse fail"`,
	}

	var tt test.TextType
	errs := Scrape(&tt, strings.NewReader(testHTML))
	if len(errs) != len(expectederrs) {
		t.Errorf("Expected %q\ngot %q", expectederrs, errs)
	} else {
		for i, err := range errs {
			if err.Error() != expectederrs[i] {
				t.Errorf("Expected %q, got %q", expectederrs[i], err.Error())
			}
		}
	}

	if tt.Struct.String1 != "11" {
		t.Errorf("Expected %q, got %q", "11", tt.Struct.String1)
	}
	if tt.Struct.String2 != "22" {
		t.Errorf("Expected %q, got %q", "22", tt.Struct.String2)
	}
	if tt.Struct.String3 != "33" {
		t.Errorf("Expected %q, got %q", "33", tt.Struct.String3)
	}
	for i, row := range tt.StructSlice {
		v1 := strings.Repeat("1", i+1)
		if row.String1 != v1 {
			t.Errorf("Expected %q, got %q", v1, row.String1)
		}
		v2 := strings.Repeat("2", i+1)
		if row.String2 != v2 {
			t.Errorf("Expected %q, got %q", v2, row.String2)
		}
		v3 := strings.Repeat("3", i+1)
		if row.String3 != v3 {
			t.Errorf("Expected %q, got %q", v3, row.String3)
		}
		markup := fmt.Sprintf("<td>%s</td><td>%s</td><td>%s</td>", v1, v2, v3)
		if row.RowMarkup != markup {
			t.Errorf("Expected %q, got %q", markup, row.RowMarkup)
		}
	}
	for i, v := range tt.Array {
		if v != i {
			t.Errorf("Expected %d, got %d", i, v)
		}
	}
	if len(tt.Slice) != 3 {
		t.Errorf("Expected slice len %d, got %d", 3, len(tt.Slice))
	}
	for i, v := range tt.Slice {
		if ev := float64(i) + 0.1; v != ev {
			t.Errorf("Expected %f, got %f", ev, v)
		}
	}
	if !bytes.Equal(tt.EightByteArray[:], []byte("EightByt")) {
		t.Errorf("Expected %q, got %q", []byte("EightByt"), tt.EightByteArray)
	}
	if !bytes.Equal(tt.ByteSlice, []byte("ByteSlice")) {
		t.Errorf("Expected %q, got %q", []byte("ByteSlice"), tt.ByteSlice)
	}
	if tt.Bool != true {
		t.Errorf("Expected %v, got %v", true, tt.Bool)
	}
	if tt.Byte != 8 {
		t.Errorf("Expected %v, got %v", 8, tt.Byte)
	}
	if tt.Int != -48 {
		t.Errorf("Expected %v, got %v", -48, tt.Int)
	}
	if tt.Int8 != 8 {
		t.Errorf("Expected %v, got %v", 8, tt.Int8)
	}
	if tt.Int16 != 16 {
		t.Errorf("Expected %v, got %v", 16, tt.Int16)
	}
	if tt.Int32 != 32 {
		t.Errorf("Expected %v, got %v", 32, tt.Int32)
	}
	if tt.Int64 != 64 {
		t.Errorf("Expected %v, got %v", 64, tt.Int64)
	}
	if *tt.Uint != 48 {
		t.Errorf("Expected %v, got %v", 48, tt.Uint)
	}
	if tt.Uint8 != 8 {
		t.Errorf("Expected %v, got %v", 8, tt.Uint8)
	}
	if tt.Uint16 != 16 {
		t.Errorf("Expected %v, got %v", 16, tt.Uint16)
	}
	if tt.Uint32 != 32 {
		t.Errorf("Expected %v, got %v", 32, tt.Uint32)
	}
	if tt.Uint64 != 64 {
		t.Errorf("Expected %v, got %v", 64, tt.Uint64)
	}
	if tt.Uintptr != 255 {
		t.Errorf("Expected %v, got %v", 255, tt.Uintptr)
	}
	if tt.Float32 != 1.234 {
		t.Errorf("Expected %v, got %v", 1.234, tt.Float32)
	}
	if tt.Float64 != 2.468 {
		t.Errorf("Expected %v, got %v", 2.468, tt.Float64)
	}
	if tt.Interface != "Interface" {
		t.Errorf("Expected %v, got %v", "Interface", tt.Interface)
	}
	if tt.String != "String" {
		t.Errorf("Expected %v, got %v", "String", tt.String)
	}
	if tt.Time.Year() != 2016 {
		t.Errorf("Expected %d, got %d", 2016, tt.Time.Year())
	}
	if tt.Time.Month() != 5 {
		t.Errorf("Expected %d, got %d", 5, tt.Time.Month())
	}
	if tt.Time.Day() != 23 {
		t.Errorf("Expected %d, got %d", 23, tt.Time.Day())
	}
	if tt.PointerToTime.Year() != 2016 {
		t.Errorf("Expected %d, got %d", 2016, tt.PointerToTime.Year())
	}
	if tt.PointerToTime.Month() != 5 {
		t.Errorf("Expected %d, got %d", 5, tt.PointerToTime.Month())
	}
	if tt.PointerToTime.Day() != 23 {
		t.Errorf("Expected %d, got %d", 23, tt.PointerToTime.Day())
	}
	if tt.URL.String() != "https://www.google.com" {
		t.Errorf("Expected: %q, got %q", "https://www.google.com", tt.URL.String())
	}
	if a := tt.Selection.Find("a"); a.Text() != "∆" {
		t.Errorf("Expected %q, got %q", "∆", a.Text())
	}
	if len(tt.Selections) != 3 {
		t.Errorf("Expected 3 items, got %d", len(tt.Selections))
	}
	for i, sel := range tt.Selections {
		if strconv.Itoa(i) != sel.Text() {
			t.Errorf("Expected %v, got %v", i, sel.Text())
		}
	}
	if tt.Node.Data != "div" {
		t.Errorf("Expected %q, got %q", "div", tt.Node.Data)
	}
	for i, node := range tt.Nodes {
		if node.Data != "p" {
			t.Errorf("Expected %q, got %q", "p", node.Data)
		}
		if node.FirstChild.Data != strconv.Itoa(i) {
			t.Errorf("Expected %d, got %q", i, node.FirstChild.Data)
		}
	}
	{
		decl := tt.Javascript.DeclarationList[0].(*ast.VariableDeclaration)
		val := decl.List[0].Initializer.(*ast.NumberLiteral)
		if val.Literal != "1" {
			t.Errorf("Expected 1, got %q", val.Literal)
		}
	}
	for i := 0; i < 3; i++ {
		decl := tt.Javascripts[i].DeclarationList[0].(*ast.VariableDeclaration)
		val := decl.List[0].Initializer.(*ast.NumberLiteral)
		if ev := strconv.Itoa(i + 1); ev != val.Literal {
			t.Errorf("Expected %q, got %q", ev, val.Literal)
		}
	}
	if val := tt.Stylesheet.Rules[0].Declarations[0].Value; val != "1px" {
		t.Errorf("Expected 1px, got %q", val)
	}
	for i := 0; i < 3; i++ {
		exp := strconv.Itoa(i+1) + "px"
		val := tt.Stylesheets[i].Rules[0].Declarations[0].Value
		if val != exp {
			t.Errorf("Expected %q, got %q", exp, val)
		}
	}

}
