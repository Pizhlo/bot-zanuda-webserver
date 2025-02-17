package model

//	{
//	  "user_id": 12345678,
//	  "text": "new note",
//	  “space_id” :1,
//	  "created": 1739264640
//	}
//
// Запрос на создание заметки
type CreateNoteRequest struct {
	UserID  int64  `json:"user_id"`  // кто создал заметку
	Text    string `json:"text"`     // текст заметки
	SpaceID int64  `json:"space_id"` // айди пространства (личного или совместного), куда сохранить заметку
	Created int64  `json:"created"`  // дата создания заметки в часовом поясе пользователя в unix
}
