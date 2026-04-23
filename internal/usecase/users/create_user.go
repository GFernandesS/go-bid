package users

import (
	"context"

	"github.com/GFernandesS/go-bid/internal/validator"
)

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (req CreateUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "user_name must not be blank")

	eval.CheckField(validator.NotBlank(req.Email), "email", "email must not be blank")

	eval.CheckField(validator.NotBlank(req.Bio), "bio", "bio must not be blank")

	eval.CheckField(validator.MinChars(req.Bio, 10) && validator.MaxChars(req.Bio, 255), "bio", "bio must have a length between 10 and 255")

	eval.CheckField(validator.MinChars(req.Password, 8), "password", "password must contain at least 8 characters")

	eval.CheckField(!validator.Matches(req.Email, validator.EmailRx), "email", "email must be a valid email")

	return eval
}
