package grammar

import (
        "testing"
        "github.com/stretchr/testify/assert"
        )

func TestGoccGrammarCreation(t *testing.T){
        assert := assert.New(t)

        _, err := NewGoccGrammar(nil)

        assert.NotNil(err)
}

func TestIsLexPartEmpty(t *testing.T){
        assert := assert.New(t)

        var empty[]byte
        var nonempty []byte

        g, err := NewGoccGrammar(empty)

        assert.NotNil(err)

        assert.Equal(g.IsLexPartEmpty(), true, "isLexPartEmpty not working, saying lexpart is not empty when it is.")

        g, err = NewGoccGrammar(nonempty)

        assert.NotNil(err)

        assert.NotEqual(g.IsLexPartEmpty(), true, "isLexPartEmpty not working, saying lexpart is not empty when it is.")
}
