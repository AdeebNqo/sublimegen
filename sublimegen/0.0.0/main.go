/*

Copyright 2014 Zola Mahlaza <adeebnqo@gmail.com>
02 December 2014

This is the driver class. It is resposible for creating the
BNF loader/parser and generating the sublime text syntax highligting
file.

*/

package main

import (
	"flag"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"os"
)

var name = flag.String("name", "default", "This is the name of the syntax.")
var scope = flag.String("scope", "source.default", "This is the scope of the syntax file")
var fileTypes = flag.String("fileTypes", "default", "Comma seperated list of file types.")

func main() {
	flag.Parse()

	jsonhighlight := `

		{ "name": "%v",
		  "scopeName": "%v",
		  "fileTypes": [%v],
		  "repository": {
		  	%v
		  },
		  "uuid": "%v"
		}
	`
	u, err := uuid.NewV4()
	if (err!=nil){
		fmt.Println("Could not generate uuid.")
		os.Exit(1)
	}else{

		repositoryfield := "COMING..."
		//0. Generate repository field from bnf file

		result := fmt.Sprintf(jsonhighlight, *name, *scope, *fileTypes, repositoryfield, u)

		//1. save result in a JSON file.

		//2. convert result to a plist file and save it.

		fmt.Println(result)
	}
}