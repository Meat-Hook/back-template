// Package repo contains wrapper for database abstraction.
package repo

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
	"github.com/Meat-Hook/back-template/libs/db"
)

var _ app.Repo = &Repo{}

type (
	// Repo provided data from and to database.
	Repo struct {
		repo *db.DB
	}

	user struct {
		ID        uuid.UUID        `db:"id"`
		Email     string           `db:"email"`
		Name      string           `db:"name"`
		PassHash  []byte           `db:"pass_hash"`
		CreatedAt pgtype.Timestamp `db:"created_at"`
		UpdatedAt pgtype.Timestamp `db:"updated_at"`
	}
)

func convert(u app.User) *user {
	return &user{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		PassHash: u.PassHash,
		CreatedAt: pgtype.Timestamp{
			Time:             u.CreatedAt,
			Status:           pgtype.Present,
			InfinityModifier: pgtype.None,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:             u.UpdatedAt,
			Status:           pgtype.Present,
			InfinityModifier: pgtype.None,
		},
	}
}

func (u user) convert() *app.User {
	return &app.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		PassHash:  u.PassHash,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

// New build and returns user db.
func New(r *db.DB) *Repo {
	return &Repo{
		repo: r,
	}
}

// Save for implements app.Repo.
func (r *Repo) Save(ctx context.Context, u app.User) (id uuid.UUID, err error) {
	err = r.repo.NoTx(func(db *sqlx.DB) error {
		newUser := convert(u)
		const query = `
		insert into 
		users 
		    (email, name, pass_hash) 
		values 
			($1, $2, $3)
		returning id
		`

		passHash := pgtype.Bytea{
			Bytes:  newUser.PassHash,
			Status: pgtype.Present,
		}

		err := db.GetContext(ctx, &id, query, newUser.Email, newUser.Name, passHash)
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", convertErr(err))
		}

		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// Update for implements app.Repo.
func (r *Repo) Update(ctx context.Context, u app.User) error {
	return r.repo.NoTx(func(db *sqlx.DB) error {
		updateUser := convert(u)

		const query = `
		update users
		set 
			email 	  = $1,
    		name  	  = $2,
    		pass_hash = $3
		where id = $4`

		passHash := pgtype.Bytea{
			Bytes:  updateUser.PassHash,
			Status: pgtype.Present,
		}

		_, err := db.ExecContext(ctx, query, updateUser.Email, updateUser.Name, passHash, updateUser.ID)
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", convertErr(err))
		}

		return nil
	})
}

// Delete for implements app.Repo.
func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `
		delete
		from users
		where id = $1`

		_, err := db.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", convertErr(err))
		}

		return nil
	})
}

// ByID for implements app.Repo.
func (r *Repo) ByID(ctx context.Context, id uuid.UUID) (u *app.User, err error) {
	err = r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `select * from users where id = $1`

		res := user{}
		err = db.GetContext(ctx, &res, query, id)
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", convertErr(err))
		}

		u = res.convert()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

// ByEmail for implements app.Repo.
func (r *Repo) ByEmail(ctx context.Context, email string) (u *app.User, err error) {
	err = r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `select * from users where email = $1`

		res := user{}
		err = db.GetContext(ctx, &res, query, email)
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", convertErr(err))
		}

		u = res.convert()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

// ByUsername for implements app.Repo.
func (r *Repo) ByUsername(ctx context.Context, username string) (u *app.User, err error) {
	err = r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `select * from users where name = $1`

		res := user{}
		err = db.GetContext(ctx, &res, query, username)
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", convertErr(err))
		}

		u = res.convert()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

// ListUserByUsername for implements app.Repo.
func (r *Repo) ListUserByUsername(ctx context.Context, username string, p app.SearchParams) (users []app.User, total int, err error) {
	err = r.repo.NoTx(func(db *sqlx.DB) error {
		const query = `SELECT * FROM users WHERE name LIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

		res := make([]user, 0, p.Limit)
		err = db.SelectContext(ctx, &res, query, "%"+username+"%", p.Limit, p.Offset)
		if err != nil {
			return fmt.Errorf("db.SelectContext: %w", convertErr(err))
		}

		const getTotal = `SELECT count(*) OVER() AS total FROM users WHERE name LIKE $1`
		err = db.GetContext(ctx, &total, getTotal, "%"+username+"%")
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", err)
		}

		users = make([]app.User, len(res))
		for i := range res {
			users[i] = *res[i].convert()
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
