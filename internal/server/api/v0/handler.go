package v0

import (
	"context"
	"errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/ex-rate/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Handler struct {
	space  spaceService
	user   userService
	auth   authService
	logger *logger.Logger

	version   string
	buildDate string
	gitCommit string
}

// интерфейс сервиса пространств. управляет пространствами, а также принадлежащими им записями: заметки, напоминания, етс
//
//go:generate mockgen -source ./handler.go -destination=./mocks/handler.go -package=mocks
type spaceService interface {
	noteCreator
	spaceCreator
	spaceChecker
	noteDeleter
	noteGetter
	spaceGetter
	noteSearcher
	noteUpdater
	participantAdder
}

type spaceCreator interface {
	CreateSpace(ctx context.Context, req rabbit.CreateSpaceRequest) error
}

type spaceChecker interface {
	IsUserInSpace(ctx context.Context, userID int64, spaceID uuid.UUID) (bool, error)
	IsSpacePersonal(ctx context.Context, spaceID uuid.UUID) (bool, error)
	IsSpaceExists(ctx context.Context, spaceID uuid.UUID) (bool, error)
	// проверяет, что приглашение от пользователя from для пользователя to в пространстве spaceID существует
	CheckInvitation(ctx context.Context, from, to int64, spaceID uuid.UUID) (bool, error)
}

type participantAdder interface {
	AddParticipant(ctx context.Context, req rabbit.AddParticipantRequest) error
}

type spaceGetter interface {
	GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error)
}

type noteCreator interface {
	CreateNote(ctx context.Context, note rabbit.CreateNoteRequest) error
}

type noteUpdater interface {
	UpdateNote(ctx context.Context, update rabbit.UpdateNoteRequest) error
}

type noteDeleter interface {
	DeleteAllNotes(ctx context.Context, req rabbit.DeleteAllNotesRequest) error
	DeleteNote(ctx context.Context, req rabbit.DeleteNoteRequest) error
}

type noteGetter interface {
	GetAllNotesBySpaceID(ctx context.Context, spaceID uuid.UUID) ([]model.GetNote, error)
	GetAllNotesBySpaceIDFull(ctx context.Context, spaceID uuid.UUID) ([]model.Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error)
	GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error)
	GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error)
}

type noteSearcher interface {
	SearchNoteByText(ctx context.Context, req model.SearchNoteByTextRequest) ([]model.GetNote, error)
}

type userService interface {
	CheckUser(ctx context.Context, tgID int64) (bool, error)
}

type authService interface {
	CheckToken(authHeader string) (*jwt.Token, error)
	GetPayload(token *jwt.Token) (jwt.MapClaims, bool)
	ParseToken(tokenString string) (*jwt.Token, error)
}

type handlerOption func(*Handler)

func WithSpaceService(space spaceService) handlerOption {
	return func(h *Handler) {
		h.space = space
	}
}

func WithUserService(user userService) handlerOption {
	return func(h *Handler) {
		h.user = user
	}
}

func WithAuthService(auth authService) handlerOption {
	return func(h *Handler) {
		h.auth = auth
	}
}

func WithLogger(logger *logger.Logger) handlerOption {
	return func(h *Handler) {
		h.logger = logger
	}
}

func WithVersion(version string) handlerOption {
	return func(h *Handler) {
		h.version = version
	}
}

func WithBuildDate(buildDate string) handlerOption {
	return func(h *Handler) {
		h.buildDate = buildDate
	}
}

func WithGitCommit(gitCommit string) handlerOption {
	return func(h *Handler) {
		h.gitCommit = gitCommit
	}
}

func New(opts ...handlerOption) (*Handler, error) {
	h := &Handler{}

	for _, opt := range opts {
		opt(h)
	}

	if h.space == nil {
		return nil, errors.New("space is nil")
	}

	if h.user == nil {
		return nil, errors.New("user is nil")
	}

	if h.auth == nil {
		return nil, errors.New("auth is nil")
	}

	if h.logger == nil {
		return nil, errors.New("logger is nil")
	}

	if h.version == "" {
		return nil, errors.New("version is nil")
	}

	if h.buildDate == "" {
		return nil, errors.New("buildDate is nil")
	}

	if h.gitCommit == "" {
		return nil, errors.New("gitCommit is nil")
	}

	h.logger.Info("handler v0 initialized")

	return h, nil
}

func (h *Handler) Stop(_ context.Context) error {
	return nil
}
