package elastic

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/deletebyquery"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
)

type ElasticIndex string

const (
	NoteIndex     ElasticIndex = "notes"
	ReminderIndex ElasticIndex = "reminders"
)

func (s ElasticIndex) String() string {
	return string(s)
}

// Структура для сохранения и поиска в ElasticSearch
type Data struct {
	Model model // Note / Reminder
	Index ElasticIndex
}

type model interface {
	validate() error
	getVal() interface{}
	searchByIDQuery() (*search.Request, error)
	searchByTextQuery() (*search.Request, error)
	deleteByQuery() (*deletebyquery.Request, error)
	updateQuery() (*update.Request, error)
	setElasticID(id string)
}

// SearchByIDQuery возвращает готовый запрос для поиска по ID.
// Ищет в эластике по ID из базы
func (d *Data) SearchByIDQuery() (*search.Request, error) {
	return d.Model.searchByIDQuery()
}

func (d *Data) SetElasticID(id string) {
	d.Model.setElasticID(id)
}

func (d *Data) DeleteByQuery() (*deletebyquery.Request, error) {
	return d.Model.deleteByQuery()
}

func (d *Data) UpdateQuery() (*update.Request, error) {
	return d.Model.updateQuery()
}

// SearchByTextQuery возвращает готовый запрос для поиска по тексту
func (d *Data) SearchByTextQuery() (*search.Request, error) {
	return d.Model.searchByTextQuery()
}

func (d *Data) ValidateNote() (*Note, error) {
	if d.Index != NoteIndex {
		return nil, fmt.Errorf("index is not equal to `notes`: `%s`", d.Index)
	}

	val := d.Model.getVal()

	note, ok := val.(Note)
	if !ok {
		return nil, fmt.Errorf("cannot convert interface{} to elastic.Note. Value: %+v", val)
	}

	return &note, d.Model.validate()
}
