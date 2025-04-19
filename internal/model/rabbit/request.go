package rabbit

import "github.com/google/uuid"

type Request struct {
	ID   uuid.UUID   `json:"id"` // айди запроса на сохранение / обновление данных
	Data interface{} // сам запрос
}
