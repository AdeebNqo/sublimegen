package token

type Token struct{
        V string
}

func (someToken *Token) IsTerminal() bool{
        return false;
}
