# hduplooy/gofixedwidth

## Fixed Width Column handling in golang

Fixed-width-column input can be read and written. It is in the same idea as encoding/csv.

It is extremely easy to use. Here follows an example of use

    package main

    import (
	    "fmt"
	    "os"
	    "strings"

	    fw "github.com/hduplooy/gofixedwidth"
    )

    func main() {
	    input := `This is a header line to be skipped
    # The following is info for the men
      John   1245
      Peter  3545
    # The following is info for certain women
      Susan  6784
      Sarah  4321
    `
	    sr := strings.NewReader(input)
	    r := fw.NewReader(sr)

	    r.Comment = '#'
	    r.SkipLines = 1
	    r.SkipStart = 2
	    r.FieldLengths = []int{7, 4}
	    tmp, _ := r.ReadAll()

	    fmt.Printf("%v\n", tmp)

	    info := [][]string{[]string{"us", "United States", "English"}, []string{"de", "Germany", "German"}, []string{"nl", "Netherlands", "Dutch"}}
	    w := fw.NewWriter(os.Stdout)
	    w.TrimFields = true
	    w.FieldLengths = []int{2, 20, 10}
	    err := w.WriteAll(info)
	    if err != nil {
		    fmt.Println(err)
		    return
	    }

    }

First a new reader is defined based on the string reader. Then it is defined that lines starting with a # are comment lines and should be skipped. A further 1 line is also skipped. 2 bytes on each line start are ignored. There are two columns of sizes 7 and 4. All of the input is then processed and a [][]string is reeturned with the data.

Next a [][]string is provided with data a new Writer is created going to standard output. If any fields are longer than defined they will be trimmed. 3 fields of length 2, 20 and 10 is defined and then all the output is send out.
