// github.com/hduplooy/gofixedwidth
// Author: Hannes du Plooy
// Revision Date: 26 Jun 2018
// Package gofixedwidth is similar to the normal encoding/csv. The difference being that the
// columns are defined with fixed widths.
// For the input the following can be defined:
//   Comment - if defined it is used to skip lines that start with this rune
//   SkipLines - the number of lines to skip before actual reading starts
//   SkipStart - indicates the number of bytes to skip on an input line before the columns are read (or to write before rest of columns are written)
//   SkipEnd - indicate how many bytes at the end of eache line to ignore (or to write after rest of columns are written)
//   TrimFields - if set all fields are trimmed (front and back) when read
//   HasEOL - indicates if lines have a CRLF or LF, or CR, when writing a CR + LF will be appended
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

const (
	EOLNONE = iota
	EOLCR
	EOLLF
	EOLCRLF
)

const (
	ALIGNLEFT = iota
	ALIGNRIGHT
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
//   HasEOL - indicates if lines have a CRLF or LF, or CR, when writing a CR + LF will be appended
//   FieldLengths - is a slice with the lengths of the fields
//	 FieldAlign - this slice contains the alignment of the field (not really of use with reading)
type Reader struct {
	Comment         rune
	SkipLines       int
	SkipStart       int
	SkipEnd         int
	FieldLengths    []int
	FieldAlign      []int
	TrimFields      bool
	HasEOL          int
	width           int
	line            int
	column          int
	initialskipdone bool
	r               *bufio.Reader
}

// readLine - read the next line from input based on the type of line delimeter (or none)

func (r *Reader) readLine() (string, error) {
	switch r.HasEOL {
	// Read up to the first CR
	case EOLCR:
		tmp, err := r.r.ReadString(13)
		if err != nil {
			return tmp, err
		}
		return tmp[:len(tmp)-1], nil

		// Read up to the first LF
	case EOLLF:
		tmp, err := r.r.ReadString(10)
		if err != nil {
			return tmp, err
		}
		return tmp[:len(tmp)-1], nil

		// Read up to the first CR and LF
	case EOLCRLF:
		tmp, err := r.r.ReadString(13)
		if err != nil {
			return tmp, err
		}
		b, err := r.r.ReadByte()
		if err != nil {
			return tmp, err
		}
		if err == nil && b == 10 {
			return tmp, nil
		}
		return tmp[:len(tmp)-1], errors.New("CRLF not found at end of line")

		// Read number of bytes based on width of fields
	case EOLNONE:
		tmp2 := make([]byte, r.width)
		_, err := r.r.Read(tmp2)
		if err != nil {
			return "", err
		}
		tmp3 := string(tmp2)
		return tmp3, nil
	}
	return "", errors.New("Nothing to return")
}

// Init updates width before everyline seeing that input
// can have different lines and thus the details can differ
func (r *Reader) Init() error {
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
	// Create a default FieldAlign if none found with all fields aligned left
	if r.FieldAlign == nil {
		r.FieldAlign = make([]int, len(r.FieldLengths))
		for i := 0; i < len(r.FieldAlign); i++ {
			r.FieldAlign[i] = ALIGNLEFT
		}
	}
	return nil
}

// NewReader returns a struct with the controls for fixed width reading
func NewReader(r io.Reader) *Reader {
	tmp := &Reader{HasEOL: EOLCRLF, r: bufio.NewReader(r)}
	tmp.Init()
	return tmp
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
	tmp, err := r.readLine()
	if err != nil {
		return nil, err
	}
	// Get rid of comment lines
	if r.Comment != 0 && rune(tmp[0]) == r.Comment {
		for rune(tmp[0]) == r.Comment {
			tmp, err = r.readLine()
			if err != nil {
				return nil, err
			}
		}
	}
	if len(tmp) != r.width {
		return nil, ErrIncorrectLineWidth
	}
	for _, val := range tmp {
		// There shouldn't be any CR or LF chars in the input
		if val == 13 || val == 10 {
			fmt.Printf("Contains cr or lf")
			return nil, ErrIncorrectLineWidth
		}
	}
	var result = make([]string, 0, len(r.FieldLengths))
	curpos := r.SkipStart                // Skip the necessary chars in beginning of line prescribed by SkipStart
	for _, val := range r.FieldLengths { // For each field extract the information
		field := string(tmp[curpos : curpos+val]) // Extract the field
		if r.TrimFields {                         // If fields must be trimmed remove any leading and trailing spaces and tabs
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
	for i := 0; i < r.SkipLines; i++ {
		_, err := r.readLine()
		if err != nil {
			return err
		}
	}
	r.initialskipdone = true
	return nil
}

// Read will read one line of fields from the input and return it
func (r *Reader) Read() ([]string, error) {
	if !r.initialskipdone {
		err := r.skipInitialLines()
		if err != nil {
			return nil, r.error(err)
		}
	}
	return r.parseRecord()
}

// ReadRows read a specified number of rows from the input
func (r *Reader) ReadRows(numOfRows int) ([][]string, error) {
	if !r.initialskipdone {
		err := r.skipInitialLines()
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
	if !r.initialskipdone {
		err := r.skipInitialLines()
		if err != nil {
			return nil, r.error(err)
		}
	}
	result := make([][]string, 0)
	for {
		record, err := r.parseRecord()
		if err != nil {
			if err.Error() == "EOF" {
				err = nil
			}
			return result, err
		}
		result = append(result, record)
	}
}

// Writer is used to control the writing to the output stream
//   Comment - if defined it is used to indicate a comment line starting with this rune
//   SkipStart - indicates the number of spaces to write before rest of columns are written)
//   SkipEnd - indicate how many spaces at the end of eache line to add
//   TrimFields - if set all fields are trimmed if they are too big else an error is returned
//   HasEOL - indicates that a CRLF must be added to each line
//   FieldLengths - is a slice with the lengths of the fields
//   FieldAlign - is a slice that contains the individual alignment of each field
type Writer struct {
	Comment      rune
	SkipStart    int
	SkipEnd      int
	FieldLengths []int
	FieldAlign   []int
	HasEOL       int
	TrimFields   bool
	width        int
	line         int
	column       int
	w            *bufio.Writer
}

// Init updates width before everyline seeing that output
// can have different lines and thus the details can differ
func (r *Writer) Init() error {
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
	// Create default alignment if none was defined
	if r.FieldAlign == nil {
		r.FieldAlign = make([]int, len(r.FieldLengths))
		for i := 0; i < len(r.FieldAlign); i++ {
			r.FieldAlign[i] = ALIGNLEFT
		}
	}
	return nil
}

// NewWriter returns a struct with the controls for fixed width writing
func NewWriter(w io.Writer) *Writer {
	tmp := &Writer{HasEOL: EOLCR, w: bufio.NewWriter(w)}
	tmp.Init()
	return tmp
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
// the output can be left aligned or right aligned and spaces will be added to accomplish this
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
			n = len(buf)
			// Add spaces in front if aligned right
			if w.FieldAlign[i] == ALIGNRIGHT {
				w.outputSpaces(w.FieldLengths[i] - n)
			}
			_, err = w.w.Write(buf)
			if err != nil {
				return err
			}
			if n != len(buf) {
				return ErrFieldLengthError
			}
			// Add spaces at back if aligned left
			if w.FieldAlign[i] == ALIGNLEFT {
				w.outputSpaces(w.FieldLengths[i] - n)
			}
		}
	}
	w.outputSpaces(w.SkipEnd)
	if w.HasEOL != EOLNONE {
		if w.HasEOL == EOLCR || w.HasEOL == EOLCRLF {
			w.w.WriteByte(13)
		}
		if w.HasEOL == EOLLF || w.HasEOL == EOLCRLF {
			w.w.WriteByte(10)
		}
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

// WriteComment send a comment character and the provided line to the output
func (w *Writer) WriteComment(line string) error {
	if w.Comment != 0 {
		_, err := w.w.WriteRune(w.Comment)
		if err != nil {
			return err
		}
		if len(line)+1 > w.width {
			line = line[0 : w.width-1]
		}
		_, err = w.w.WriteString(line)
		if len(line)+1 < w.width {
			w.outputSpaces(w.width - len(line) - 1)
		}
		if err != nil {
			return err
		}
		// Output line delimeter if defined
		if w.HasEOL != EOLNONE {
			if w.HasEOL == EOLCR || w.HasEOL == EOLCRLF {
				w.w.WriteByte(13)
			}
			if w.HasEOL == EOLLF || w.HasEOL == EOLCRLF {
				w.w.WriteByte(10)
			}
		}
	}
	return nil
}
