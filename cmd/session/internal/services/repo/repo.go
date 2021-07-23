// Package repo contains implements for app.Repo.
// Provide session info to and from repository.
package repo

import (
	"context"
	"fmt"
	"net"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/libs/db"
)

var _ app.Repo = &Repo{}

type (
	// Repo provided data from and to database.
	Repo struct {
		repo *db.DB
	}

	session struct {
		ID        pgtype.UUID      `db:"id"`
		Token     string           `db:"token"`
		IP        pgtype.Inet      `db:"ip"`
		UserAgent string           `db:"user_agent"`
		UserID    pgtype.UUID      `db:"user_id"`
		CreatedAt pgtype.Timestamp `db:"created_at"`
		UpdatedAt pgtype.Timestamp `db:"updated_at"`
	}
)

// New build and returns session db.
func New(r *db.DB) *Repo {
	return &Repo{
		repo: r,
	}
}

func convert(s app.Session) (*session, error) {
	ip, err := inet(s.Origin.IP)
	if err != nil {
		return nil, fmt.Errorf("parse ip: %w", err)
	}

	return &session{
		ID: pgtype.UUID{
			Bytes:  s.ID,
			Status: pgtype.Present,
		},
		Token:     s.Token.Value,
		IP:        *ip,
		UserAgent: s.Origin.UserAgent,
		UserID: pgtype.UUID{
			Bytes:  s.UserID,
			Status: pgtype.Present,
		},
		CreatedAt: pgtype.Timestamp{
			Time:             s.CreatedAt,
			Status:           pgtype.Present,
			InfinityModifier: pgtype.None,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:             s.UpdatedAt,
			Status:           pgtype.Present,
			InfinityModifier: pgtype.None,
		},
	}, nil
}

func (s session) convert() *app.Session {
	return &app.Session{
		ID: s.ID.Bytes,
		Origin: app.Origin{
			IP:        s.IP.IPNet.IP,
			UserAgent: s.UserAgent,
		},
		Token: app.Token{
			Value: s.Token,
		},
		UserID:    s.UserID.Bytes,
		CreatedAt: s.CreatedAt.Time,
		UpdatedAt: s.UpdatedAt.Time,
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
			return nil, fmt.Errorf("inet.Set: %w", err)
		}
	}

	return inet, nil
}

// Save for implements app.Repo.
func (r *Repo) Save(ctx context.Context, session app.Session) error {
	return r.repo.NoTx(func(db *sqlx.DB) error {
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

		_, err = db.NamedExecContext(ctx, query, newSession)
		if err != nil {
			return fmt.Errorf("db.NamedExecContext: %w", err)
		}

		return nil
	})
}

// ByID for implements app.Repo.
func (r *Repo) ByID(ctx context.Context, sessionID uuid.UUID) (s *app.Session, err error) {
	err = r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `select * from sessions where id = $1`

		res := session{}
		err = db.GetContext(ctx, &res, query, sessionID)
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", convertErr(err))
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
func (r *Repo) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `
		delete
		from sessions
		where id = $1`

		_, err := db.ExecContext(ctx, query, sessionID)
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", err)
		}

		return nil
	})
}
