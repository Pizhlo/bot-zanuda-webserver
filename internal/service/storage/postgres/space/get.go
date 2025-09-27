package space

import (
	"context"
	"database/sql"
	"errors"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (db *Repo) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
	res := model.Space{}

	err := db.db.QueryRowContext(ctx, "select id, name, created, creator, personal from shared_spaces.shared_spaces where id = $1", id).
		Scan(&res.ID, &res.Name, &res.Created, &res.Creator, &res.Personal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Space{}, api_errors.ErrSpaceNotExists
		}
	}

	logrus.WithField("id", id).Debug("got space from postgres by ID")

	return res, err
}

func (db *Repo) IsSpacePersonal(ctx context.Context, spaceID uuid.UUID) (bool, error) {
	var personal bool
	err := db.db.QueryRowContext(ctx, "select personal from shared_spaces.shared_spaces where id = $1", spaceID).
		Scan(&personal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, api_errors.ErrSpaceNotExists
		}

		logrus.WithField("spaceID", spaceID).Debug("got space not personal from postgres by ID")

		return false, err
	}

	logrus.WithField("spaceID", spaceID).Debug("got space personal from postgres by ID")

	return personal, nil
}

func (db *Repo) IsSpaceExists(ctx context.Context, spaceID uuid.UUID) (bool, error) {
	var exists bool
	err := db.db.QueryRowContext(ctx, "select exists(select 1 from shared_spaces.shared_spaces where id = $1)", spaceID).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	logrus.WithField("spaceID", spaceID).Debug("got space exists from postgres by ID")

	return exists, nil
}

func (db *Repo) CheckInvitation(ctx context.Context, from, to int64, spaceID uuid.UUID) (bool, error) {
	var exists bool
	err := db.db.QueryRowContext(ctx, `select exists(select 1 from shared_spaces.invitations 
where "from" = (select id from users.users where tg_id = $1)
and "to" = (select id from users.users where tg_id = $2)
and space_id = $3);`, from, to, spaceID).
		Scan(&exists)
	if err != nil {
		logrus.WithField("spaceID", spaceID).Debug("got invitation not exists from postgres by ID")
		return false, err
	}

	logrus.WithField("from", from).WithField("to", to).WithField("spaceID", spaceID).Debug("got invitation exists from postgres by ID")

	return exists, nil
}

// CheckParticipant проверяет, является ли пользователь участником пространства
func (db *Repo) CheckParticipant(ctx context.Context, userID int64, spaceID uuid.UUID) (bool, error) {
	logrus.WithField("userID", userID).WithField("spaceID", spaceID).Debug("checking participant")

	space, err := db.GetSpaceByID(ctx, spaceID) // получаем информацию о пространстве (проверить, личное ли оно)
	if err != nil {
		return false, err
	}

	if space.Personal {
		var userSpaceID uuid.UUID

		// выясняем айди личного пространства пользователя
		err := db.db.QueryRowContext(ctx, "select space_id from users.users where tg_id = $1", userID).
			Scan(&userSpaceID)
		if err != nil {
			return false, err
		}

		if userSpaceID == spaceID {
			return true, nil
		}

		return false, nil
	}

	var exists bool

	err = db.db.QueryRowContext(ctx, `select exists(select 1 from shared_spaces.participants 
where user_id = (select id from users.users where tg_id = $1) 
and space_id = $2
and state_id = 2)`, userID, spaceID).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		return true, nil
	}

	var creator, tgID int64
	// выясняем, не является ли пользователь создателем пространства
	err = db.db.QueryRowContext(ctx, `select creator, tg_id from shared_spaces.shared_spaces
join users.users on users.users.id = shared_spaces.shared_spaces.creator
where  shared_spaces.shared_spaces.id = $1`, spaceID).Scan(&creator, &tgID) // creator - айди в БД, tgID - айди в телеге
	if err != nil {
		return false, err
	}

	if userID == tgID {
		return true, nil
	}

	return false, nil
}
