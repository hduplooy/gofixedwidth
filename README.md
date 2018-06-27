# hduplooy/gofixedwidth

## Fixed Width Column handling in golang

Fixed-width-column input can be read and written. It is in the same idea as encoding/csv.

It is extremely easy to use. Here follows an example of use

package main

	import (
		"fmt"
		"strings"
	
		ar "github.com/hduplooy/goarrrecords"
		fw "github.com/hduplooy/gofixedwidth"
	)
	
	type Person struct {
		Name string
		Id   int
	}
	
	type Employee struct {
		EmpID int
		Id    int
		Name  string
	}
	
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
	
		r.Comment = '#'              // Skip any lines staritng with '#'
		r.SkipLines = 1              // Skip the very first line
		r.SkipStart = 2              // Skip first 2 characters on the input line
		r.FieldLengths = []int{7, 4} // First field is 7 characters long and next one is 4
		r.HasEOL = fw.EOLLF          // The lines are delimited with LF
		r.TrimFields = true
		r.Init() // Just make sure everything is ready for reading
		tmp, err := r.ReadAll()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Records from LF delimited input\nlength=", len(tmp))
		fmt.Println(tmp)
	
		// Reading data from fixed width with no CR or LF
		input2 := `Skip me      # Some commnt  Ivan   4444  Carin  8934`
		sr = strings.NewReader(input2)
		r = fw.NewReader(sr)
		r.Comment = '#'
		r.SkipLines = 1
		r.SkipStart = 2
		r.FieldLengths = []int{7, 4}
		r.HasEOL = fw.EOLNONE
		r.Init()
		tmp1, err := r.ReadAll()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Records with no line delimited input\nlength=", len(tmp1))
		fmt.Println(tmp1)
	
		var recs []Person
	
		tmp2, err := ar.Arr2Records(tmp, Person{})
		if err != nil {
			fmt.Println(err)
			return
		}
		recs = tmp2.([]Person)
		fmt.Printf("%v\n", recs)
		tmp3, err := ar.Records2Arr(recs)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%v\n", tmp3)
		emps, err := ar.CopyRecs(Employee{}, recs)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%v\n", emps)
	
		sw := strings.Builder{}
		w := fw.NewWriter(&sw)
		w.HasEOL = fw.EOLLF
		w.SkipStart = 2              // Skip first 2 characters on the input line
		w.FieldLengths = []int{7, 4} // First field is 7 characters long and next one is 4
		w.Init()
		w.WriteAll(tmp)
	
		fmt.Println(sw.String())
	}


First a new reader is defined based on the string reader. Then it is defined that lines starting with a # are comment lines and should be skipped. A further 1 line is also skipped. 2 bytes on each line start are ignored. There are two columns of sizes 7 and 4. All of the input is then processed and a [][]string is returned with the data.

Next a [][]string is provided with data a new Writer is created going to standard output. If any fields are longer than defined they will be trimmed. 3 fields of length 2, 20 and 10 is defined and then all the output is send out.
