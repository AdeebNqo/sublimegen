package main

import (
	"fmt"
	"os"
)
func main() {
	var myfile *os.File
	var err error
	myfile, err = os.Open("textfile.txt")
	if (err != nil){
		fmt.Println("Were in trouble.")
	}else{
		data := make([]byte, 100)

		myfile.Read(data)
		fmt.Println("----\tdata\t----")
		fmt.Println(string(data))
	}
}