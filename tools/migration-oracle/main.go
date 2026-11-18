// Command migration-oracle captures the observable legacy dec128 contract.
// It intentionally uses only the published Badassitron API, so it can be built
// independently from either the original module or a candidate module.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	d "github.com/profe-ajedrez/badassitron/dec128"
	derr "github.com/profe-ajedrez/badassitron/dec128/errors"
)

type Case struct {
	ID      string `json:"id"`
	N       int    `json:"n,omitempty"`
	Op      string `json:"op"`
	A       string `json:"a,omitempty"`
	B       string `json:"b,omitempty"`
	Scale   uint8  `json:"scale,omitempty"`
	Mode    uint8  `json:"mode,omitempty"`
	SrcType string `json:"src_type,omitempty"`
}

type Result struct {
	ID          string  `json:"id"`
	Op          string  `json:"op"`
	String      string  `json:"string"`
	Fixed       string  `json:"fixed"`
	Scale       uint8   `json:"scale"`
	IsNaN       bool    `json:"nan"`
	IsZero      bool    `json:"zero"`
	IsNegative  bool    `json:"negative"`
	Error       string  `json:"error,omitempty"`
	Text        string  `json:"text"`
	JSON        string  `json:"json"`
	SQLValue    string  `json:"sql_value"`
	SQLError    string  `json:"sql_error,omitempty"`
	Inexact     *bool   `json:"inexact,omitempty"`
	ScanError   string  `json:"scan_error,omitempty"`
	Extra       *Result `json:"extra,omitempty"`
	Encoded     string  `json:"encoded,omitempty"`
	EncodeError string  `json:"encode_error,omitempty"`
}

func run(c Case) (r Result, err error) {
	r.ID, r.Op = c.ID, c.Op
	a, b := d.FromString(c.A), d.FromString(c.B)
	var out d.Dec128
	switch c.Op {
	case "parse":
		out = a
	case "add":
		out = a.Add(b)
	case "sub":
		out = a.Sub(b)
	case "mul":
		out = a.Mul(b)
	case "div":
		out = a.Div(b)
	case "compare":
		out = d.FromInt(a.Compare(b))
	case "mod":
		out = a.Mod(b)
	case "quorem":
		var remainder d.Dec128
		out, remainder = a.QuoRem(b)
		extra, e := capture(Result{ID: c.ID, Op: "remainder"}, remainder)
		if e != nil {
			return r, e
		}
		r.Extra = &extra
	case "powint":
		out = a.PowInt(c.N)
	case "encode_uint128":
		out = a
		encoded, e := out.EncodeToUint128(c.Scale)
		if e != nil {
			r.EncodeError = e.Error()
		} else {
			r.Encoded = encoded.String()
		}
	case "canonical":
		out = a.Canonical()
	case "round":
		out = a.Round(c.Scale)
	case "roundbank":
		out = a.RoundBank(c.Scale)
	case "rescale":
		out = a.Rescale(c.Scale)
	case "mulround":
		out = a.MulRound(b, c.Scale, d.RoundingMode(c.Mode))
	case "divround":
		out = a.DivRound(b, c.Scale, d.RoundingMode(c.Mode))
	case "mulstatus":
		var inexact bool
		out, inexact = a.MulRoundWithStatus(b, c.Scale, d.RoundingMode(c.Mode))
		r.Inexact = &inexact
	case "divstatus":
		var inexact bool
		out, inexact = a.DivRoundWithStatus(b, c.Scale, d.RoundingMode(c.Mode))
		r.Inexact = &inexact
	case "nan":
		code, convErr := strconv.ParseUint(c.A, 10, 8)
		if convErr != nil {
			return r, convErr
		}
		out = d.NaN(derr.Error(code))
	case "scan":
		out = d.FromString("9.99")
		var src any
		switch c.SrcType {
		case "string":
			src = c.A
		case "bytes":
			src = []byte(c.A)
		case "int64":
			src, err = strconv.ParseInt(c.A, 10, 64)
		case "float64":
			src, err = strconv.ParseFloat(c.A, 64)
		case "nil":
			src = nil
		case "unsupported":
			src = true
		default:
			return r, fmt.Errorf("unknown src_type %q", c.SrcType)
		}
		if err != nil {
			return r, err
		}
		if e := d.ScanInto(&out, src); e != nil {
			r.ScanError = e.Error()
		}
	case "unmarshaljson":
		out = d.FromString("9.99")
		if e := out.UnmarshalJSON([]byte(c.A)); e != nil {
			r.ScanError = e.Error()
		}
	case "unmarshaltext":
		out = d.FromString("9.99")
		if e := out.UnmarshalText([]byte(c.A)); e != nil {
			r.ScanError = e.Error()
		}
	default:
		return r, fmt.Errorf("unknown operation %q", c.Op)
	}
	return capture(r, out)
}

func capture(r Result, out d.Dec128) (Result, error) {
	r.String, r.Fixed = out.String(), out.StringFixed()
	r.Scale, r.IsNaN, r.IsZero, r.IsNegative = out.Precision(), out.IsNaN(), out.IsZero(), out.IsNegative()
	if e := out.ErrorDetails(); e != nil {
		r.Error = e.Error()
	}
	txt, e := out.MarshalText()
	if e != nil {
		return r, e
	}
	r.Text = string(txt)
	jb, e := out.MarshalJSON()
	if e != nil {
		return r, e
	}
	r.JSON = string(jb)
	val, e := out.Value()
	if e != nil {
		r.SQLError = e.Error()
	} else {
		r.SQLValue = fmt.Sprint(val)
	}
	return r, nil
}

func main() {
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 64*1024), 1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	line := 0
	for in.Scan() {
		line++
		raw := bytes.TrimSpace(in.Bytes())
		if len(raw) == 0 {
			continue
		}
		var c Case
		if err := json.Unmarshal(raw, &c); err != nil {
			fmt.Fprintf(os.Stderr, "line %d: %v\n", line, err)
			os.Exit(1)
		}
		r, err := run(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "line %d (%s): %v\n", line, c.ID, err)
			os.Exit(1)
		}
		payload, err := json.Marshal(r)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintln(out, string(payload))
	}
	if err := in.Err(); err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
