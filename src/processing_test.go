package main

import (
  "testing"
  "io/ioutil"
  )

func TestStartsAndEndWithQuotation(t *testing.T){
	correctString := "\"hello\""
  incorrectStrnig := "hello"
  if (!startsAndEndWithQuotation(correctString)){
      t.Errorf("startsAndEndWithQuotation() does not work for correct strings");
  }
  if (startsAndEndWithQuotation(incorrectStrnig)){
    t.Errorf("startsAndEndWithQuotation() does not work for incorrect strings");
  }
}

func TestDoesDirExist(t *testing.T){
  tempdir,err := ioutil.TempDir(".","")
  if (err!=nil){
    t.Error(err);
  }
  if (!doesDirExist(&tempdir)){
    t.Errorf("doesDirExist() does not detect folder that exists.");
  }
}

func TestDeleteDir(t *testing.T){
  tempdir,err := ioutil.TempDir(".","")
  if (err!=nil){
    t.Error(err)
  }
  if (deleteDir(&tempdir)!=nil){
    t.Errorf("deleteDir() does not work. Cannot delete file.");
  }
}
func TestConvertJSONtoPlist(t *testing.T){
  //creating tmp json file named: test.tmLanguage.json
  jsontext := []byte("{\"value\": \"New\", \"onclick\": \"CreateNewDoc()\"}")
  err := ioutil.WriteFile("test.tmLanguage.json", jsontext, 0775)
  if (err!=nil){
    t.Error(err)
  }
  filename := "test"
  err = convertJSONtoPlist(&filename)
  if (err!=nil){
    t.Error(err)
  }
}

func TestStartandendwithrb(t *testing.T){
  correctString := "(hello)"
  incorrectString := "hello"
  if (!startandendwithrb(correctString)){
    t.Errorf("startandendwithrb() does not work. Doesn't detect strings that start with round braces");
  }
  if (startandendwithrb(incorrectString)){
    t.Errorf("startandendwithrb() does not work. Says strings that do not start with round braces do.");
  }
}
