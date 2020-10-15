package jl

import (
	"fmt"
	"io"
	"strings"
)

// ULFPrinter can print logs in a variety of compact formats, specified by FieldFormats.
type ULFPrinter struct {
	Out io.Writer
	// Disable colors disables adding color to fields.
	DisableColor bool
	// Disable truncate disables the Ellipsize and Truncate transforms.
	DisableTruncate bool
	// FieldFormats specifies the format the printer should use for logs. It defaults to DefaultCompactPrinterFieldFmt. Fields
	// are formatted in the order they are provided. If a FieldFmt produces a field that does not end with a whitespace,
	// a space character is automatically appended.
	ServiceFieldFormats       []FieldFmt
	CommunicationFieldFormats []FieldFmt
}

// DefaultULFServicePrinterFieldFmt is a format for the ULFPrinter that tries to present logs in an easily skimmable manner
// for most types of logs.
var DefaultULFServicePrinterFieldFmt = []FieldFmt{{
	Name:         "level",
	Finders:      []FieldFinder{ByNames("level")},
	Transformers: []Transformer{Truncate(4), UpperCase, ColorMap(LevelColors)},
}, {
	Name:    "time",
	Finders: []FieldFinder{ByNames("timestamp")},
}, {
	Name:         "instance",
	Finders:      []FieldFinder{ByNames("application.instanceId")},
	Transformers: []Transformer{Ellipsize(18), Format("[%s]"), RightPad(20), ColorSequence(AllColors)},
}, {
	Name:         "logger",
	Finders:      []FieldFinder{ByNames("application.component")},
	Transformers: []Transformer{Compress(30), Format("%s"), LeftPad(30), ColorSequence(AllColors)},
}, {
	Name:    "description",
	Finders: []FieldFinder{ByNames("details.description")},
}, {
	Name:    "details",
	Finders: []FieldFinder{ByNames("details.details")},
}}

// DefaultULFCommunicationPrinterFieldFmt is a format for the ULFPrinter that tries to present logs in an easily skimmable manner
// for most types of logs.
var DefaultULFCommunicationPrinterFieldFmt = []FieldFmt{{
	Name:         "level",
	Finders:      []FieldFinder{ByNames("level")},
	Transformers: []Transformer{Truncate(4), UpperCase, ColorMap(LevelColors)},
}, {
	Name:    "time",
	Finders: []FieldFinder{ByNames("timestamp")},
}, {
	Name:         "instance",
	Finders:      []FieldFinder{ByNames("application.instanceId")},
	Transformers: []Transformer{Ellipsize(18), Format("[%s]"), RightPad(20), ColorSequence(AllColors)},
}, {
	Name:         "logger",
	Finders:      []FieldFinder{ByNames("application.component")},
	Transformers: []Transformer{Compress(30), Format("%s"), LeftPad(30), ColorSequence(AllColors)},
}, {
	Name:    "flow",
	Finders: []FieldFinder{ByNames("details.flow")},
}, {
	Name:    "method",
	Finders: []FieldFinder{ByNames("details.request.method")},
}, {
	Name:    "url",
	Finders: []FieldFinder{ByNames("details.request.url")},
}, {
	Name:    "statusCode",
	Finders: []FieldFinder{ByNames("details.response.statusCode")},
}}

// NewCompactPrinter allocates and returns a new compact printer.
func NewULFPrinter(w io.Writer) *ULFPrinter {
	return &ULFPrinter{
		Out:                       w,
		ServiceFieldFormats:       DefaultULFServicePrinterFieldFmt,
		CommunicationFieldFormats: DefaultULFCommunicationPrinterFieldFmt,
	}
}

func (p *ULFPrinter) Print(entry *Entry) {
	if entry.Partials == nil {
		fmt.Fprintln(p.Out, string(entry.Raw))
		return
	}
	ctx := Context{
		DisableColor:    p.DisableColor,
		DisableTruncate: p.DisableTruncate,
	}
	category := DefaultStringer(&ctx, entry.Partials["category"])
	var fieldFormats []FieldFmt
	if category == "SERVICE" {
		p.Out.Write([]byte("SERV "))
		fieldFormats = p.ServiceFieldFormats
	} else if category == "COMMUNICATION" {
		p.Out.Write([]byte("COMM "))
		fieldFormats = p.CommunicationFieldFormats
	} else {
		fmt.Fprintln(p.Out, string(entry.Raw))
		return
	}
	for i, fieldFmt := range fieldFormats {
		formattedField := fieldFmt.format(&ctx, entry)
		if formattedField != "" {
			if i != 0 && !strings.HasPrefix(formattedField, "\n") {
				p.Out.Write([]byte(" "))
			}
			p.Out.Write([]byte(formattedField))
		}
	}
	p.Out.Write([]byte("\n"))
}
