package rabbit

import (
	"webserver/internal/model"

	"github.com/google/uuid"
)

type CreateSpaceRequest struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Created   int64     `json:"created"`
	UserID    int64     `json:"user_id"` // создатель пространства
	Operation Operation `json:"operation"`
}

func (c CreateSpaceRequest) Validate() error {
	if c.ID == uuid.Nil {
		return model.ErrIDNotFilled
	}

	if len(c.Name) == 0 {
		return model.ErrFieldNameNotFilled
	}

	if c.Created == 0 {
		return model.ErrFieldCreatedNotFilled
	}

	if c.Operation != CreateOp {
		return ErrInvalidOperation
	}

	if c.UserID == 0 {
		return model.ErrFieldUserNotFilled
	}

	return nil
}
