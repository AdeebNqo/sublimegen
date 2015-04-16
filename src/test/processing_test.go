package main
import (
        "testing"
        )
func TestStripliteral(t *testing.T){
        value := "\"hello\"";
        result := stripliteral(value);
        if (result!="hello"){
                t.Errorf("Stripliteral does not work.");
        }
}

func TestStartandendwithrb(t *testing.T){
        wrongValue := "hello"
        rightValue := "(hello)"
        if (startandendwithrb(wrongValue)){
                t.Errorf("Startandendwithrb is failing for wrong value.");
        }
        if (!startandendwithrb(rightValue)){
                t.Errorf("Startandendwithrb is failing for right value.");
        }
}

/*func TestEscape(t *testing.T){
        bslash = "/"
        dollar = "$"
        rsbrace = "["
        lsbrace = "]"
        fslash = "\\"
        space = " "
        plus = "+"
        lrbrace = "("
        rrbrace = ")"
        lcbrace = "{"
        rcbrace = "}"
        questionmark = "?"
        leftdash = "`"
        dot = "."
}*/
