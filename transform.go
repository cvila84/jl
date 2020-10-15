package jl

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Context provides the current transformation context, to be used by Transformers and Stringers.
type Context struct {
	// The original string before any transformations were applied.
	Original string
	// Indicates that terminal color escape sequences should be disabled.
	DisableColor bool
	// Indicates that fields should not be truncated.
	DisableTruncate bool
}

// Transformer transforms a string and returns the result.
type Transformer interface {
	Transform(ctx *Context, input string) string
}

// TransformFunc is an adapter to allow the use of ordinary functions as Transformers.
type TransformFunc func(string) string

func (f TransformFunc) Transform(ctx *Context, input string) string {
	return f(input)
}

var (
	// UpperCase transforms the input string to upper case.
	UpperCase = TransformFunc(strings.ToUpper)
	// LowerCase transforms the input string to lower case.
	LowerCase = TransformFunc(strings.ToLower)
)

// Compress compresses a dot tokenized string in reducing left side token(s) to one letter so that it fits within the
// requested length. (See http://logback.qos.ch/manual/layouts.html#conversionWord)
type Compress int

func (c Compress) Transform(ctx *Context, input string) string {
	if ctx.DisableTruncate {
		return input
	}
	compressedSize := utf8.RuneCountInString(input)
	if compressedSize <= int(c) {
		return input
	}
	compressedPartsIndex := 0
	parts := strings.Split(input, ".")
	nCompressableParts := len(parts) - 1
	var outputBuffer strings.Builder
	if nCompressableParts > 0 {
		var compressedParts []string
		for _, part := range parts[:nCompressableParts] {
			compressedPartsIndex++
			compressedParts = append(compressedParts, string(part[0]))
			compressedSize -= len(part) - 1
			if compressedSize <= int(c) {
				break
			}
			compressedSize--
		}
		for _, part := range compressedParts {
			outputBuffer.WriteString(part)
		}
		outputBuffer.WriteRune('.')
		if compressedPartsIndex < nCompressableParts {
			for _, part := range parts[compressedPartsIndex:nCompressableParts] {
				outputBuffer.WriteString(part)
			}
			outputBuffer.WriteRune('.')
		}
	}
	lastPart := parts[nCompressableParts]
	if int(c)-outputBuffer.Len()-len(lastPart) >= 0 {
		outputBuffer.WriteString(lastPart)
	} else if int(c)-outputBuffer.Len() >= 0 {
		outputBuffer.WriteString(lastPart[:int(c)-outputBuffer.Len()])
	} else {
		outputBuffer.Reset()
		if len(lastPart) > int(c) {
			outputBuffer.WriteString(lastPart[:c])
		} else {
			outputBuffer.WriteString(lastPart)
		}
	}
	return outputBuffer.String()
}

// Truncate truncates the string to the a requested number of digits.
type Truncate int

func (t Truncate) Transform(ctx *Context, input string) string {
	if ctx.DisableTruncate {
		return input
	}
	if utf8.RuneCountInString(input) <= int(t) {
		return input
	}
	return input[:t]
}

// Ellipsize replaces characters in the middle of the string with a single "…" character so that it fits within the
// requested length.
type Ellipsize int

func (remain Ellipsize) Transform(ctx *Context, input string) string {
	if ctx.DisableTruncate {
		return input
	}
	length := utf8.RuneCountInString(input)
	if length <= int(remain) {
		return input
	}
	remain -= 1 // account for the ellipsis
	chomped := length - int(remain)
	start := int(remain) / 2
	end := start + chomped
	return input[:start] + "…" + input[end:]
}

// LeftPad pads the left side of the string with spaces so that the string becomes the requested length.
type LeftPad int

func (t LeftPad) Transform(ctx *Context, input string) string {
	spaces := int(t) - utf8.RuneCountInString(input)
	if spaces <= 0 {
		return input
	}
	buf := bytes.NewBuffer(make([]byte, 0, spaces+len(input)))
	for i := 0; i < spaces; i++ {
		buf.WriteRune(' ')
	}
	buf.WriteString(input)
	return buf.String()
}

// LeftPad pads the right side of the string with spaces so that the string becomes the requested length.
type RightPad int

func (t RightPad) Transform(ctx *Context, input string) string {
	pad := int(t) - utf8.RuneCountInString(input)
	if pad <= 0 {
		return input
	}
	buf := bytes.NewBuffer(make([]byte, 0, pad+len(input)))
	buf.WriteString(input)
	for i := 0; i < pad; i++ {
		buf.WriteRune(' ')
	}
	return buf.String()
}

// Format calls fmt.Sprintf() with the requested format string.
type Format string

func (t Format) Transform(ctx *Context, input string) string {
	return fmt.Sprintf(string(t), input)
}
