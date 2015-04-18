package production

import(
        "github.com/AdeebNqo/sublimegen/src/token"
        )
type Production struct{
        leftHand string
        rightHand []token.Token
}

func (someProduction *Production) ReadLeftHand() string{
        return someProduction.leftHand
}
func (someProduction *Production) ReadRightHand() interface{}{
        return someProduction.rightHand
}
