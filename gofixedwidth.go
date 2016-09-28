// github.com/hduplooy/gofixedwidth
// Author: Hannes du Plooy
// Revision Date: 28 Sep 2016
// Package gofixedwidth is similar to the normal encoding/csv. The difference being that the
// columns are defined with fixed widths.
// For the input the following can be defined:
//   Comment - if defined it is used to skip lines that start with this rune
//   SkipLines - the number of lines to skip before actual reading starts
//   SkipStart - indicates the number of bytes to skip on an input line before the columns are read (or to write before rest of columns are written)
//   SkipEnd - indicate how many bytes at the end of eache line to ignore (or to write after rest of columns are written)
//   TrimFields - if set all fields are trimmed (front and back) when read
//   HasEOL - indicates if lines have a CRLF (or LF, or CR), when writing a CR + LF will be appended
//   FieldLengths - is a slice with the lengths of the fields
//
// For each line a slice of strings are returned when read
package gofixedwidth

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Used to generate any errors experienced
type ParseError struct {
	Line   int
	Column int
	Err    error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Err)
}

var (
	ErrFieldCount         = errors.New("wrong number of fields in line")
	ErrNoFields           = errors.New("no fields defined to read")
	ErrFieldLengthError   = errors.New("fields width incorrect")
	ErrIncorrectLineWidth = errors.New("incorrect line width")
	ErrNotEnoughLines     = errors.New("not enough lines")
)

// Reader is used to control the reading from the input stream
//   Comment - if defined it is used to skip lines that start with this rune
//   SkipLines - the number of lines to skip before actual reading starts
//   SkipStart - indicates the number of bytes to skip on an input line before the columns are read (or to write before rest of columns are written)
//   SkipEnd - indicate how many bytes at the end of eache line to ignore (or to write after rest of columns are written)
//   TrimFields - if set all fields are trimmed (front and back) when read
//   HasEOL - indicates if lines have a CRLF (or LF, or CR), when writing a CR + LF will be appended
//   FieldLengths - is a slice with the lengths of the fields
type Reader struct {
	Comment         rune
	SkipLines       int
	SkipStart       int
	SkipEnd         int
	FieldLengths    []int
	TrimFields      bool
	HasEOL          bool
	width           int
	line            int
	column          int
	initialskipdone bool
	r               *bufio.Reader
}

// updateWidth updates width before everyline seeing that input
// can have different lines and thus the details can differ
func (r *Reader) updateWidth() error {
	if r.SkipStart < 0 {
		r.SkipStart = 0
	}
	if r.SkipEnd < 0 {
		r.SkipEnd = 0
	}
	r.width = r.SkipStart + r.SkipEnd
	if len(r.FieldLengths) == 0 {
		return ErrNoFields
	}
	for _, val := range r.FieldLengths {
		if val <= 0 {
			return ErrFieldLengthError
		}
		r.width += val
	}
	return nil
}

// NewReader returns a struct with the controls for fixed width reading
func NewReader(r io.Reader) *Reader {
	return &Reader{HasEOL: true, r: bufio.NewReader(r)}
}

// error generates a ParseError with necessary information
func (r *Reader) error(err error) error {
	return &ParseError{Line: r.line, Column: r.column, Err: err}
}

// parseRecord process a line
// First any lines with comments (if comment is defined) are skipped
// The number of bytes based on the width is then read.
// If it either is too small or contains a CR or LF an error is returned (because it means the line length is incorrect).
// If HasEOL is defined and no CR/LF follows it means there are extra characters on the line which is an error
// Then based on the field lengths the fields are extracted and trimmed (if defined).
func (r *Reader) parseRecord() (fields []string, err error) {
	var tmp = make([]byte, r.width)
	if r.HasEOL && r.Comment != '\000' {
		b, _, err := r.r.ReadRune()
		if err != nil {
			return nil, err
		}
		for b == r.Comment {
			_, _, err = r.r.ReadLine()
			if err != nil {
				return nil, err
			}
			b, _, err = r.r.ReadRune()
			if err != nil {
				return nil, err
			}
		}
		r.r.UnreadRune()
	}
	n, err := r.r.Read(tmp)
	if err != nil {
		return nil, err
	}
	if n != r.width {
		return nil, ErrIncorrectLineWidth
	}
	for _, val := range tmp {
		if val == 13 || val == 10 {
			fmt.Printf("Contains cr or lf")
			return nil, ErrIncorrectLineWidth
		}
	}
	if r.HasEOL {
		b, err := r.r.ReadByte()
		if err != nil {
			return nil, err
		}
		if b != 13 && b != 10 {
			return nil, ErrIncorrectLineWidth
		}
		if b == 10 {
			b, err := r.r.ReadByte()
			if b != 13 && err == nil {
				r.r.UnreadByte()
			}
		}
	}

	var result = make([]string, 0, len(r.FieldLengths))
	curpos := r.SkipStart
	for _, val := range r.FieldLengths {
		field := string(tmp[curpos : curpos+val])
		if r.TrimFields {
			field = strings.Trim(field, " \t")
		}
		curpos += val
		result = append(result, field)
	}
	return result, nil
}

// skipInitialLines - will only be called once after the definition of Reader
// it will skip the number of lines defined (if defined)
func (r *Reader) skipInitialLines() error {
	if r.HasEOL {
		for i := 0; i < r.SkipLines; i++ {
			_, _, err := r.r.ReadLine()
			if err != nil {
				return err
			}
		}
	}
	r.initialskipdone = true
	return nil
}

// Read will read one line of fields from the input and return it
func (r *Reader) Read() ([]string, error) {
	err := r.updateWidth()
	if err != nil {
		return nil, r.error(err)
	}
	if !r.initialskipdone {
		err = r.skipInitialLines()
		if err != nil {
			return nil, r.error(err)
		}
	}
	return r.parseRecord()
}

// ReadRows read a specified number of rows from the input
func (r *Reader) ReadRows(numOfRows int) ([][]string, error) {
	err := r.updateWidth()
	if err != nil {
		return nil, r.error(err)
	}
	if !r.initialskipdone {
		err = r.skipInitialLines()
		if err != nil {
			return nil, r.error(err)
		}
	}
	result := make([][]string, 0, numOfRows)
	for i := 0; i < numOfRows; i++ {
		record, err := r.parseRecord()
		if err != nil {
			return result, r.error(err)
		}
		result = append(result, record)
	}
	return result, nil
}

// ReadAll will read all lines from the input
func (r *Reader) ReadAll() ([][]string, error) {
	err := r.updateWidth()
	if err != nil {
		return nil, r.error(err)
	}
	if !r.initialskipdone {
		err = r.skipInitialLines()
		if err != nil {
			return nil, r.error(err)
		}
	}
	result := make([][]string, 0)
	for {
		record, err := r.parseRecord()
		if err != nil {
			return result, r.error(err)
		}
		result = append(result, record)
	}
}

// Writer is used to control the writing to the output stream
//   SkipStart - indicates the number of spaces to write before rest of columns are written)
//   SkipEnd - indicate how many spaces at the end of eache line to add
//   TrimFields - if set all fields are trimmed if they are too big else an error is returned
//   HasEOL - indicates that a CRLF must be added to each line
//   FieldLengths - is a slice with the lengths of the fields
type Writer struct {
	SkipStart    int
	SkipEnd      int
	FieldLengths []int
	HasEOL       bool
	TrimFields   bool
	width        int
	line         int
	column       int
	w            *bufio.Writer
}

// updateWidth updates width before everyline seeing that output
// can have different lines and thus the details can differ
func (r *Writer) updateWidth() error {
	if r.SkipStart < 0 {
		r.SkipStart = 0
	}
	if r.SkipEnd < 0 {
		r.SkipEnd = 0
	}
	r.width = r.SkipStart + r.SkipEnd
	if len(r.FieldLengths) == 0 {
		return ErrNoFields
	}
	for _, val := range r.FieldLengths {
		if val <= 0 {
			return ErrFieldLengthError
		}
		r.width += val
	}
	return nil
}

// NewWriter returns a struct with the controls for fixed width writing
func NewWriter(w io.Writer) *Writer {
	return &Writer{HasEOL: true, w: bufio.NewWriter(w)}
}

// error generates a ParseError with necessary information
func (w *Writer) error(err error) error {
	return &ParseError{Line: w.line, Column: w.column, Err: err}
}

// outputSpaces will send a specific number of spaces to the output
func (w *Writer) outputSpaces(n int) {
	for n > 0 {
		w.w.WriteByte(' ')
		n--
	}
}

// Write will first output the defined number of spaces at the front (SkipStart)
// then the fields are output (trimmed if need be) and then any trailing spaces are added (if SkipEnd is defined)
// If HasEOL is defined CR and LF will be send to output
func (w *Writer) Write(flds []string) error {
	w.outputSpaces(w.SkipStart)
	if len(flds) != len(w.FieldLengths) {
		return ErrFieldCount
	}
	for i := 0; i < len(flds); i++ {
		buf := []byte(flds[i])
		var n int
		var err error
		if len(buf) > w.FieldLengths[i] {
			if !w.TrimFields {
				return ErrFieldLengthError
			}
			n, err = w.w.Write(buf[0:w.FieldLengths[i]])
			if err != nil {
				return err
			}
			if n != w.FieldLengths[i] {
				return ErrFieldLengthError
			}
		} else {
			n, err = w.w.Write(buf)
			if err != nil {
				return err
			}
			if n != len(buf) {
				return ErrFieldLengthError
			}
			w.outputSpaces(w.FieldLengths[i] - n)
		}
	}
	w.outputSpaces(w.SkipEnd)
	if w.HasEOL {
		w.w.WriteByte(13)
		w.w.WriteByte(10)
	}
	return nil
}

// WriteAll will write every record in the slice to output
func (w *Writer) WriteAll(recs [][]string) error {
	for _, record := range recs {
		err := w.Write(record)
		if err != nil {
			return err
		}
	}
	w.w.Flush()
	return nil
}

// Flush will flush the output stream
func (w *Writer) Flush() {
	w.w.Flush()
}
