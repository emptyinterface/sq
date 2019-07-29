package test

import (
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aymerick/douceur/css"
	"github.com/robertkrimen/otto/ast"
	"golang.org/x/net/html"
)

type (
	Row struct {
		private   int
		String1   string `sq:"td:nth-child(1) | text"`
		String2   string `sq:"td:nth-child(2) | text"`
		String3   string `sq:"td:nth-child(3) | text"`
		RowMarkup string `sq:" . | html"`
	}
	CustomType string
	TextType   struct {
		private        string
		privatepointer *string
		Struct         Row                  `sq:"table.list tr:nth-child(2)"`
		StructSlice    []*Row               `sq:"table.list tr"`
		Array          [3]int               `sq:"p.array | text"`
		Slice          []float64            `sq:"p.slice | text | regexp([.\\d]+)"`
		ByteSlice      []byte               `sq:"p.byteslice | text"`
		EightByteArray [8]byte              `sq:"p.eightbytearray | text"`
		Bool           bool                 `sq:"p.bool | text"`
		Byte           byte                 `sq:"p.byte | text"`
		Int            int                  `sq:"p.int | text"`
		Int8           int8                 `sq:"p.int8 | text"`
		Int16          int16                `sq:"p.int16 | text"`
		Int32          int32                `sq:"p.int32 | text"`
		Int64          int64                `sq:"p.int64 | text"`
		Uint           *uint                `sq:"p.uint | text"`
		Uint8          uint8                `sq:"p.uint8 | text"`
		Uint16         uint16               `sq:"p.uint16 | text"`
		Uint32         uint32               `sq:"p.uint32 | text"`
		Uint64         uint64               `sq:"p.uint64 | text"`
		Uintptr        uintptr              `sq:"p.uintptr | text"`
		Float32        float32              `sq:"p.float32 | text"`
		Float64        float64              `sq:"p.float64 | text"`
		Interface      interface{}          `sq:"p.interface | text"`
		String         string               `sq:"p.string | text"`
		Time           time.Time            `sq:"p.time | text | regexp([\\d\\s]{10,}) | time(2006 01 02)"`
		PointerToTime  *time.Time           `sq:"p.time | text | regexp([\\d\\s]{10,}) | time(2006 01 02)"`
		URL            *url.URL             `sq:"a | attr(href)"`
		Selection      *goquery.Selection   `sq:"div"`
		Selections     []*goquery.Selection `sq:"div > p"`
		Node           *html.Node           `sq:"div"`
		Nodes          []*html.Node         `sq:"div > p"`
		Javascript     *ast.Program         `sq:"script[type$=javascript]:first-child"`
		Javascripts    []*ast.Program       `sq:"script[type$=javascript]"`
		Stylesheet     *css.Stylesheet      `sq:"style:first-of-type"`
		Stylesheets    []*css.Stylesheet    `sq:"style"`
		CustomType     CustomType           `sq:"p.string"`
		Optional       string               `sq:"(optional) blink"`

		// errs
		Map                 map[string]interface{} `sq:"div"`
		BadBool             bool                   `sq:"p.int | text"`
		BadInt              int                    `sq:"p.bool | text"`
		BadUint             uint                   `sq:"p.bool | text"`
		BadFloat            float32                `sq:"p.bool | text"`
		BadTime             time.Time              `sq:"p.bool | text | time()"`
		BadSlice            []byte                 `sq:"div | attr(missing)"`
		BadArray            [8]byte                `sq:"div | attr(missing)"`
		BadAttr             int                    `sq:"div | attr(missing)"`
		BadTag              int                    `sq:"derp(\d)"`
		BadParse            string                 `sq:"p.bool | text | parsefail"`
		BadLoad             string                 `sq:"p.bool | text | loadfail"`
		BadSliceofStructs   []Badstruct            `sq:"div"`
		BadArrayofStructs   [2]Badstruct           `sq:"div"`
		privatetagged       string                 `sq:"a"`
		Missing             string                 `sq:"blink"`
		MissingSelection    *goquery.Selection     `sq:"blink.selection"`
		MissingNode         *html.Node             `sq:"blink.node"`
		MissingJavascript   *ast.Program           `sq:"blink.javascript"`
		MissingStylesheet   *css.Stylesheet        `sq:"blink.css"`
		BadAccSelection     *goquery.Selection     `sq:"a | badacc.goquery"`
		BadAccNode          *html.Node             `sq:"a | badacc.node"`
		BadAccURL           *url.URL               `sq:"a | badacc.url"`
		BadAccJavascript    *ast.Program           `sq:"a | badacc.javascript"`
		BadAccStylesheet    *css.Stylesheet        `sq:"a | badacc.css"`
		BadParserSelection  *goquery.Selection     `sq:"a | text | parsefail"`
		BadParserNode       *html.Node             `sq:"a | text | parsefail"`
		BadParserURL        *url.URL               `sq:"a | text | parsefail"`
		BadParserJavascript *ast.Program           `sq:"a | text | parsefail"`
		BadParserStylesheet *css.Stylesheet        `sq:"a | text | parsefail"`
	}
	Badstruct struct {
		Field string `sq:"div | text | nestedfail"`
	}
)
