package main
import (
        "testing"
        "github.com/stretchr/testify/assert"
        )
func TestStripliteral(t *testing.T){
        assert := assert.New(t)

        value := "\"hello\"";
        result := stripliteral(value);

        assert.Equal(result, "hello", "Stripliteral does not work.")
}

func TestStartandendwithrb(t *testing.T){
        assert := assert.New(t)

        wrongValue := "hello"
        rightValue := "(hello)"

        assert.Equal(startandendwithrb(wrongValue), false, "Startandendwithrb is failing for wrong value.");

        assert.Equal(startandendwithrb(rightValue), true, "Startandendwithrb is failing for right value.")
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
