package gofixedwidth

/*
PACKAGE DOCUMENTATION

package gofixedwidth
    import "github.com/hduplooy/gofixedwidth"

    github.com/hduplooy/gofixedwidth Author: Hannes du Plooy Revision Date:
    26 Jun 2018 Package gofixedwidth is similar to the normal encoding/csv.
    The difference being that the columns are defined with fixed widths. For
    the input the following can be defined:

	Comment - if defined it is used to skip lines that start with this rune
	SkipLines - the number of lines to skip before actual reading starts
	SkipStart - indicates the number of bytes to skip on an input line before the columns are read (or to write before rest of columns are written)
	SkipEnd - indicate how many bytes at the end of eache line to ignore (or to write after rest of columns are written)
	TrimFields - if set all fields are trimmed (front and back) when read
	HasEOL - indicates if lines have a CRLF or LF, or CR, when writing a CR + LF will be appended
	FieldLengths - is a slice with the lengths of the fields

    For each line a slice of strings are returned when read

CONSTANTS

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

VARIABLES

var (
    ErrFieldCount         = errors.New("wrong number of fields in line")
    ErrNoFields           = errors.New("no fields defined to read")
    ErrFieldLengthError   = errors.New("fields width incorrect")
    ErrIncorrectLineWidth = errors.New("incorrect line width")
    ErrNotEnoughLines     = errors.New("not enough lines")
)

TYPES

type ParseError struct {
    Line   int
    Column int
    Err    error
}
    Used to generate any errors experienced

func (e *ParseError) Error() string

type Reader struct {
    Comment      rune
    SkipLines    int
    SkipStart    int
    SkipEnd      int
    FieldLengths []int
    FieldAlign   []int
    TrimFields   bool
    HasEOL       int
    // contains filtered or unexported fields
}
    Reader is used to control the reading from the input stream

	  Comment - if defined it is used to skip lines that start with this rune
	  SkipLines - the number of lines to skip before actual reading starts
	  SkipStart - indicates the number of bytes to skip on an input line before the columns are read (or to write before rest of columns are written)
	  SkipEnd - indicate how many bytes at the end of eache line to ignore (or to write after rest of columns are written)
	  TrimFields - if set all fields are trimmed (front and back) when read
	  HasEOL - indicates if lines have a CRLF or LF, or CR, when writing a CR + LF will be appended
	  FieldLengths - is a slice with the lengths of the fields
		 FieldAlign - this slice contains the alignment of the field (not really of use with reading)

func NewReader(r io.Reader) *Reader
    NewReader returns a struct with the controls for fixed width reading

func (r *Reader) Init() error
    Init updates width before everyline seeing that input can have different
    lines and thus the details can differ

func (r *Reader) Read() ([]string, error)
    Read will read one line of fields from the input and return it

func (r *Reader) ReadAll() ([][]string, error)
    ReadAll will read all lines from the input

func (r *Reader) ReadRows(numOfRows int) ([][]string, error)
    ReadRows read a specified number of rows from the input

type Writer struct {
    Comment      rune
    SkipStart    int
    SkipEnd      int
    FieldLengths []int
    FieldAlign   []int
    HasEOL       int
    TrimFields   bool
    // contains filtered or unexported fields
}
    Writer is used to control the writing to the output stream

	Comment - if defined it is used to indicate a comment line starting with this rune
	SkipStart - indicates the number of spaces to write before rest of columns are written)
	SkipEnd - indicate how many spaces at the end of eache line to add
	TrimFields - if set all fields are trimmed if they are too big else an error is returned
	HasEOL - indicates that a CRLF must be added to each line
	FieldLengths - is a slice with the lengths of the fields
	FieldAlign - is a slice that contains the individual alignment of each field

func NewWriter(w io.Writer) *Writer
    NewWriter returns a struct with the controls for fixed width writing

func (w *Writer) Flush()
    Flush will flush the output stream

func (r *Writer) Init() error
    Init updates width before everyline seeing that output can have
    different lines and thus the details can differ

func (w *Writer) Write(flds []string) error
    Write will first output the defined number of spaces at the front
    (SkipStart) the output can be left aligned or right aligned and spaces
    will be added to accomplish this then the fields are output (trimmed if
    need be) and then any trailing spaces are added (if SkipEnd is defined)
    If HasEOL is defined CR and LF will be send to output

func (w *Writer) WriteAll(recs [][]string) error
    WriteAll will write every record in the slice to output

func (w *Writer) WriteComment(line string) error
    WriteComment send a comment character and the provided line to the
    output

SUBDIRECTORIES

	examples
*/
