package schema_postgres

import (
	"fmt"

	"github.com/Xrefullx/golang-shorturl/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type (
	User struct {
		ID uuid.UUID `validate:"required"`
	}
)

func NewUserFromCanonical(obj model.User) (User, error) {
	dbObj := User{
		ID: obj.ID,
	}
	if err := dbObj.Validate(); err != nil {
		return User{}, err
	}
	return dbObj, nil
}

func (u User) ToCanonical() (model.User, error) {
	obj := model.User{
		ID: u.ID,
	}

	if err := obj.Validate(); err != nil {
		return model.User{}, fmt.Errorf("status: %w", err)
	}

	return obj, nil
}

func (u User) Validate() error {
	validate := validator.New()

	if err := validate.Struct(u); err != nil {
		return fmt.Errorf("error validation db User : %w", err)
	}

	return nil
}
