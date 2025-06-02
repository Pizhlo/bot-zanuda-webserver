package rabbit

import "errors"

type Operation string

var (
	CreateOp         Operation = "create"
	UpdateOp         Operation = "update"
	DeleteOp         Operation = "delete"
	DeleteAllOp      Operation = "delete_all"
	AddParticipantOp Operation = "add_participant"
)

var (
	ErrInvalidOperation = errors.New("invalid operation")
)
