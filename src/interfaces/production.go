package interfaces

import(
        "github.com/AdeebNqo/sublimegen/src/model"
        )
type Production interface{
        ReadLeftHand() model.Token
        ReadRighthand() []model.Token
}
