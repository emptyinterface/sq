# sq

sq is a very simple, powerful scraping library

sq uses [struct tags](https://golang.org/pkg/reflect/#StructTag) as configuration, [reflection](golang.org/pkg/reflect), and [goquery](https://github.com/PuerkitoBio/goquery) to unmarshall data out of HTML pages.

```go
type ExamplePage struct {
	Title string `sq:"title | text"`

	Users []struct {
		ID        int       `sq:"td:nth-child(1) | text | regexp(\\d+)"`
		Name      string    `sq:"td:nth-child(2) | text"`
		Email     string    `sq:"td:nth-child(3) a | attr(href) | regexp(mailto:(.+))"`
		Website   *url.URL  `sq:"td:nth-child(4) > a | attr(href)"`
		Timestamp time.Time `sq:"td:nth-child(5) | text | time(2006 02 03)"`
		RowMarkup string    `sq:" . | html"`
	} `sq:"table tr"`

	Stylesheets []*css.Stylesheet `sq:"style"`
	Javascripts []*ast.Program    `sq:"script [type$=javascript]"`

	HTMLSnippet      *html.Node         `sq:"div.container"`
	GoquerySelection *goquery.Selection `sq:"[href], [src]"`
}

resp, err := http.Get("https://example.com")
if err != nil {
	log.Fatal(err)
}
defer resp.Body.Close()

var p ExamplePage

// Scrape continues on error and returns a slice of errors that occurred.
errs := sq.Scrape(&p, resp.Body)
for _, err := range errs {
	fmt.Println(err)
}
```


*Note: go struct tags are parsed as strings and so all backslashes must be escaped.  (ie. `\d+` -> `\\d+`)*

## Accessors, Parsers, and Loaders

Accessors, parsers, loaders are specified in the tag in a unix-style pipeline.

 **Accessors**

  * `text`: The `text` accessor emits the result of goquery's [`Text()`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection.Text) method on the matched [`Selection`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection).
  * `html`: The `html` accessor emits the result of goquery's [`Html()`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection.Html) method on the matched [`Selection`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection).
  * `attr(<attr>)`: The `attr()` accessor emits the result of goquery's [`Attr()`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection.Attr) method with the supplied argument on the matched [`Selection`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection).  An error will be returned if the specified attribute is not found.

**Parsers**

 * `regexp(<regexp>)`:  The `regexp` parser takes a regular expression and applies it to the input emitted by the previous accessor or parser function.  When no subcapture group is specified, the first match is emitted.  If a subcapture group is specified, the first subcapture is returned.

**Loaders**

 * `time(<format>)`:  The `time()` loader calls [`time.Parse()`](https://golang.org/pkg/time/#Parse) with the supplied format on the input emitted from the previous accessor or parser function.

Custom parsers and loaders may be added or overridden:

```go
// unescapes content
sq.RegisterParseFunc("unescape", func(s, _ string) (string, error) {
	return html.UnescapeString(s), nil
})

// loads a time.Duration from a datestamp
sq.RegisterLoadFunc("age", func(_ *goquery.Selection, s, layout string) (interface{}, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return nil, err
	}
	return time.Since(t), nil
})

// example use
type Page struct {
	Alerts []struct {
		Title string        `sq:"h3 | text"`
		Age   time.Duration `sq:"span.posted | unescape | age(2006 02 03 15:04:05 MST)"`
	} `sq:"div.alert"`
}
```


## Types

sq supports the full list of native go types except `map`, `func`, `chan`, and `complex`.

Several web related datastructures are also detected and loaded:

 * [`url.URL`](https://golang.org/pkg/net/url/#URL):  The `url.URL` type from the go std lib loaded using [`url.Parse`](https://golang.org/pkg/net/url/#Parse)
 * [`github.com/aymerick/douceur/css.Stylesheet`](https://godoc.org/github.com/aymerick/douceur/css#Stylesheet): This is a parse tree representing a stylesheet.
 * [`github.com/robertkrimen/otto/ast.Program`](https://godoc.org/github.com/robertkrimen/otto/ast#Program): This is an ast representing a block of javascript.
 * [`golang.org/x/net/html.Node`](https://godoc.org/golang.org/x/net/html#Node): This is the ast node of the parsed html.
 * [`github.com/PuerkitoBio/goquery.Selection`](https://godoc.org/github.com/PuerkitoBio/goquery#Selection): This is a convenience wrapper around the underlying html node[s].

Each of these types are detected and loaded automatically using a [`TypeLoader`](https://godoc.org/github.com/emptyinterface/sq#TypeLoader).  Overriding or adding type loaders is simple.

A [`TypeLoader`](https://godoc.org/github.com/emptyinterface/sq#TypeLoader) is a pair of functions with a name.  It takes function that checks for a match, and a function that does the loading.

```go
// This is the typeloader for detecting url.URLs and loading them.
sq.RegisterTypeLoader("url",
	func(t reflect.Type) bool {
		return t.PkgPath() == "net/url" && t.Name() == "URL"
	},
	func(_ *goquery.Selection, s string) (interface{}, error) {
		return url.Parse(s)
	},
)
```

### Docs

[godoc](https://godoc.org/github.com/emptyinterface/sq)

### License

MIT 2016

