package users

import (
	"context"

	"github.com/GFernandesS/go-bid/internal/validator"
)

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginUserRequest) Valid(ctx context.Context) validator.Evaluator {

	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.Email), "email", "must be a valid email")

	eval.CheckField(validator.NotBlank(req.Password), "password", "must be a valid password")

	return eval
}
