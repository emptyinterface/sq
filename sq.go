package sq

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

var (
	// reflection errors
	ErrInvalidKind       = errors.New("invalid kind")
	ErrNotSettable       = errors.New("v is not settable")
	ErrNonStructPtrValue = errors.New("*struct type required")
	ErrTagNotFound       = errors.New("sq tag not found")

	// not found errors
	ErrNodeNotFound      = errors.New("node not found")
	ErrAttributeNotFound = errors.New("attribute not found")
)

func Scrape(structPtr interface{}, r io.Reader) []error {

	v := reflect.ValueOf(structPtr)

	if v.Kind() != reflect.Ptr || v.Type().Elem().Kind() != reflect.Struct {
		return []error{ErrNonStructPtrValue}
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return []error{err}
	}

	return hydrateValue(&v, doc.Selection, nil)

}

// initialize and dereference pointers
func resolvePointer(v *reflect.Value) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		*v = v.Elem()
	}
}

func hydrateValue(v *reflect.Value, sel *goquery.Selection, p *path) []error {

	resolvePointer(v)

	if !v.CanSet() {
		return nil
	}

	if p != nil && len(p.selector) > 0 && p.selector != "." && !sel.Is(p.selector) {
		sel = sel.Find(p.selector)
		if sel.Size() == 0 {
			return []error{fmt.Errorf("%q did not match", p.selector)}
		}
	}

	if p != nil && p.loader != nil {
		if err := setValueFromSel(v, sel, p); err != nil {
			return []error{err}
		}
		return nil
	}

	t := v.Type()

	for _, tl := range typeLoaders {
		if tl.isType(t) {
			p.loader = &loader{
				f: func(sel *goquery.Selection, text, _ string) (interface{}, error) {
					return tl.load(sel, text)
				},
			}
			if err := setValueFromSel(v, sel, p); err != nil {
				return []error{err}
			}
			return nil
		}
	}

	switch v.Kind() {

	case reflect.Struct:

		var errs []error
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			p, err := parseTag(ft.Tag)
			if err != nil {
				if err != ErrTagNotFound {
					errs = append(errs, err)
				}
			} else {
				if r, _ := utf8.DecodeRuneInString(ft.Name); !unicode.IsUpper(r) {
					errs = append(errs, fmt.Errorf("private field with sq tag: %q", ft.Name))
				} else {
					f := v.Field(i)
					if err := hydrateValue(&f, sel, p); err != nil {
						errs = append(errs, err...)
					}
				}
			}
		}
		return errs

	case reflect.Array:

		// handle [N]byte copy from string
		if t.Elem().Kind() == reflect.Uint8 {
			s, err := extractString(sel, p.acc)
			if err != nil {
				return []error{err}
			}
			reflect.Copy(*v, reflect.ValueOf([]byte(s)))
			return nil
		}

		var errs []error
		sel.Each(func(i int, sel *goquery.Selection) {
			if i < v.Len() {
				vv := v.Index(i)
				if err := hydrateValue(&vv, sel, p); err != nil {
					errs = append(errs, err...)
				}
			}
		})
		return errs

	case reflect.Slice:

		// handle []byte setting directly
		if t.Elem().Kind() == reflect.Uint8 {
			s, err := extractString(sel, p.acc)
			if err != nil {
				return []error{err}
			}
			v.SetBytes([]byte(s))
			return nil
		}

		var errs []error
		slicev := reflect.MakeSlice(t, sel.Size(), sel.Size())
		sel.Each(func(i int, sel *goquery.Selection) {
			vv := slicev.Index(i)
			if err := hydrateValue(&vv, sel, p); err != nil {
				errs = append(errs, err...)
			}
		})
		v.Set(slicev)
		return errs

	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Interface,
		reflect.String:
		if err := setValueFromSel(v, sel, p); err != nil {
			return []error{err}
		}
		return nil

	default:
		// case reflect.Map:
		// case reflect.Complex64:
		// case reflect.Complex128:
		// case reflect.Chan:
		// case reflect.Func:
		return []error{fmt.Errorf("%s: %v", ErrInvalidKind, v.Kind())}
	}

}

func setValueFromSel(v *reflect.Value, sel *goquery.Selection, p *path) error {

	s, err := extractString(sel, p.acc)
	if err != nil {
		return err
	}

	for _, pp := range p.parsers {
		s, err = pp.parse(s)
		if err != nil {
			return fmt.Errorf("%s: (parser fail) %q", p.selector, err)
		}
	}

	if p.loader != nil {
		vv, err := p.loader.load(sel, s)
		if err != nil {
			return fmt.Errorf("%s: (loader fail) %q", p.selector, err)
		}
		rv := reflect.ValueOf(vv)
		// deref pointers for correct setting.
		// by this point we're operating on
		// non-pointer values only.
		for rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		v.Set(rv)
		return nil
	}

	switch v.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("%s: %s", p.selector, err)
		}
		v.SetBool(b)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		n, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("%s: %s", p.selector, err)
		}
		v.SetInt(n)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("%s: %s", p.selector, err)
		}
		v.SetUint(n)
	case reflect.Float32,
		reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("%s: %s", p.selector, err)
		}
		v.SetFloat(n)
	case reflect.String:
		v.SetString(s)
	case reflect.Interface:
		v.Set(reflect.ValueOf(s))
	default:
		// unable to set these values from a string
		// but should never reach this block with
		// one of these value kinds.
		// case reflect.Slice:
		// case reflect.Struct:
		// case reflect.Array:
		// case reflect.Map:
		// case reflect.Complex64:
		// case reflect.Complex128:
		// case reflect.Chan:
		// case reflect.Func:
		panic("unreachable")
	}

	return nil

}
