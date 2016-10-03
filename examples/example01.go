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

	r.Comment = '#'
	r.SkipLines = 1
	r.SkipStart = 2
	r.FieldLengths = []int{7, 4}
	tmp, err := r.ReadAll()

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
}
