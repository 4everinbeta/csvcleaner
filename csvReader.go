package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Reader interface {
	Read() ([]string, error)
}

type Value interface {
	String() string
	Set(string) bool
}

type ReadIter struct {
	Reader       Reader
	Headers      []string
	Error        error
	Line, Column int
	fields       []reflect.Value
	kinds        []int
	tags         []int
}

const (
	none_k = iota
	string_k
	int_k
	float_k
	uint_k
	value_k
)

// Creates a new iterator from a Reader source and a user-defined struct.
func NewReadIter(file *os.File, ps interface{}) (this *ReadIter, err error) {
	rdr := csv.NewReader(file)
	// options are available at:
	// http://golang.org/src/pkg/encoding/csv/reader.go?s=3213:3671#L94
	rdr.Comma = ','

	this = new(ReadIter)
	this.Line = 1
	this.Headers, err = rdr.Read()
	this.Reader = rdr
	if err != nil {
		this = nil
		return
	}
	st := reflect.TypeOf(ps).Elem()
	sv := reflect.ValueOf(ps).Elem()
	nf := st.NumField()
	// if nf != len(this.Headers) {
	// 	return nil, &FieldMismatch{nf, len(this.Headers)}
	// }
	this.kinds = make([]int, nf)
	this.tags = make([]int, nf)
	this.fields = make([]reflect.Value, nf)
	for i := 0; i < nf; i++ {
		f := st.Field(i)

		val := sv.Field(i)
		// get the corresponding field name and look it up in the headers
		tag := f.Tag.Get("field")
		if len(tag) == 0 {
			tag = f.Name
			if strings.Contains(tag, "_") {
				tag = strings.Replace(tag, "_", " ", -1)
			}
		}

		reqTag := f.Tag.Get("required")
		fmt.Printf("isRequired: %v\n", reqTag)

		isRequired, parseErr := strconv.ParseBool(reqTag)
		if parseErr != nil {
			err = errors.New("Unable to parse required tag " + reqTag)
		}

		itag := -1
		for k, h := range this.Headers {
			if h == tag {
				itag = k
				break
			}
		}

		if itag == -1 && isRequired {
			err = errors.New("cannot find this field " + tag)
			this = nil
			return
		}

		kind := none_k
		Kind := f.Type.Kind()
		// this is necessary because Kind can't tell distinguish between a primitive type
		// and a type derived from it. We're looking for a Value interface defined on
		// the pointer to this value
		_, ok := val.Addr().Interface().(Value)
		if ok {
			val = val.Addr()
			kind = value_k
		} else {
			switch Kind {
			case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int32, reflect.Int64:
				kind = int_k
			case reflect.Uint, reflect.Uint16, reflect.Uint8, reflect.Uint32, reflect.Uint64:
				kind = uint_k
			case reflect.Float32, reflect.Float64:
				kind = float_k
			case reflect.String:
				kind = string_k
			default:
				kind = value_k
				_, ok := val.Interface().(Value)
				if !ok {
					err = errors.New("cannot convert this type ")
					this = nil
					return
				}
			}
		}
		if itag >= 0 {
			this.kinds[i] = kind
			this.tags[i] = itag
			this.fields[i] = val
		}

	}
	return
}

// The Get method reads the next row. If there was an error or EOF, it
// will return false.  Client code must then check that ReadIter.Error is
// not nil to distinguish between normal EOF and specific errors.
func (this *ReadIter) Get() bool {
	row, err := this.Reader.Read()
	this.Line = this.Line + 1
	if err != nil {
		if err != io.EOF {
			this.Error = err
		}
		return false
	}
	var ival int64
	var fval float64
	var uval uint64
	var v Value
	var ok bool

	for fi, ci := range this.tags {
		vals := row[ci] // string at column ci of current row
		f := this.fields[fi]
		switch this.kinds[fi] {
		case string_k:
			f.SetString(vals)
		case int_k:
			ival, err = strconv.ParseInt(vals, 10, 64)
			f.SetInt(ival)
		case uint_k:
			uval, err = strconv.ParseUint(vals, 10, 64)
			f.SetUint(uval)
		case float_k:
			fval, err = strconv.ParseFloat(vals, 64)
			f.SetFloat(fval)
		case value_k:
			v, ok = f.Interface().(Value)
			if !ok {
				err = errors.New("Not a Value object")
				break
			}
			v.Set(vals)
		}
		if err != nil {
			this.Column = ci + 1
			this.Error = err
			return false
		}
	}
	return true
}

// Error types
type FieldMismatch struct {
	expected, found int
}

func (e *FieldMismatch) Error() string {
	return "CSV line fields mismatch. Expected " + strconv.Itoa(e.expected) + " found " + strconv.Itoa(e.found)
}

type UnsupportedType struct {
	Type string
}

func (e *UnsupportedType) Error() string {
	return "Unsupported type: " + e.Type
}
