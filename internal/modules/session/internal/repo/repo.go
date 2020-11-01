package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
)

var _ app.Repo = &Repo{}

type (
	// Repo provided data from and to database.
	Repo struct {
		db     *sqlx.DB
		metric *metrics.Database
	}

	session struct {
		ID        string      `db:"id"`
		Token     string      `db:"token"`
		IP        pgtype.Inet `db:"ip"`
		UserAgent string      `db:"user_agent"`
		UserID    int         `db:"user_id"`
		CreatedAt time.Time   `db:"created_at"`
		UpdatedAt time.Time   `db:"updated_at"`
	}
)

// New build and returns user repo.
func New(db *sqlx.DB, m *metrics.Database) *Repo {
	return &Repo{
		db:     db,
		metric: m,
	}
}

func convert(s app.Session) (*session, error) {
	ip, err := inet(s.Origin.IP)
	if err != nil {
		return nil, fmt.Errorf("parse ip: %w", err)
	}

	return &session{
		ID:        s.ID,
		Token:     s.Token.Value,
		IP:        *ip,
		UserAgent: s.Origin.UserAgent,
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}

func (s session) convert() *app.Session {
	return &app.Session{
		ID: s.ID,
		Origin: app.Origin{
			IP:        s.IP.IPNet.IP,
			UserAgent: s.UserAgent,
		},
		Token: app.Token{
			Value: s.Token,
		},
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func inet(ip net.IP) (*pgtype.Inet, error) {
	inet := &pgtype.Inet{}
	if ip == nil || ip.IsUnspecified() {
		err := inet.Set(nil)
		if err != nil {
			return nil, err
		}
	} else {
		err := inet.Set(ip)
		if err != nil {
			return nil, fmt.Errorf("inet set: %w", err)
		}
	}

	return inet, nil
}

// Save for implements app.Repo.
func (r *Repo) Save(ctx context.Context, session app.Session) error {
	return r.metric.Collect(func() error {
		newSession, err := convert(session)
		if err != nil {
			return fmt.Errorf("convert session: %w", err)
		}

		const query = `
		insert into 
		sessions 
		    (id, token, ip, user_agent, user_id) 
		values 
			(:id, :token, :ip, :user_agent, :user_id)
		`

		_, err = r.db.NamedExecContext(ctx, query, newSession)
		if err != nil {
			return fmt.Errorf("save session: %w", err)
		}

		return nil
	})
}

// ByID for implements app.Repo.
func (r *Repo) ByID(ctx context.Context, sessionID string) (s *app.Session, err error) {
	err = r.metric.Collect(func() error {
		const query = `select * from sessions where id = $1`

		res := session{}
		err = r.db.GetContext(ctx, &res, query, sessionID)
		if err != nil {
			return fmt.Errorf("get session by id: %w", convertErr(err))
		}

		s = res.convert()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Delete for implements app.Repo.
func (r *Repo) Delete(ctx context.Context, sessionID string) error {
	return r.metric.Collect(func() error {
		const query = `
		delete
		from sessions
		where id = $1`

		_, err := r.db.ExecContext(ctx, query, sessionID)
		if err != nil {
			return fmt.Errorf("delete session: %w", err)
		}

		return nil
	})
}

func convertErr(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return app.ErrNotFound
	default:
		return err
	}
}
