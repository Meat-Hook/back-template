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
		ID        pgtype.UUID      `db:"id"`
		Email     string           `db:"email"`
		Name      string           `db:"name"`
		PassHash  pgtype.Bytea     `db:"pass_hash"`
		Avatars   pgtype.UUIDArray `db:"avatars"`
		CreatedAt pgtype.Timestamp `db:"created_at"`
		UpdatedAt pgtype.Timestamp `db:"updated_at"`
	}
)

func convert(u app.User) *user {
	idStatus := pgtype.Present
	if u.ID == uuid.Nil {
		idStatus = pgtype.Null
	}

	id := pgtype.UUID{
		Bytes:  u.ID,
		Status: idStatus,
	}

	passHashStatus := pgtype.Present
	if len(u.PassHash) == 0 {
		passHashStatus = pgtype.Null
	}

	passHash := pgtype.Bytea{
		Bytes:  u.PassHash,
		Status: passHashStatus,
	}

	avatarsUUID := pgtype.UUIDArray{
		Elements:   make([]pgtype.UUID, 0, len(u.Avatars)),
		Dimensions: nil,
		Status:     pgtype.Present,
	}

	for i := range u.Avatars {
		fileIDStatus := pgtype.Present
		if u.Avatars[i] == uuid.Nil {
			fileIDStatus = pgtype.Null
		}

		fileID := pgtype.UUID{
			Bytes:  u.Avatars[i],
			Status: fileIDStatus,
		}

		avatarsUUID.Elements = append(avatarsUUID.Elements, fileID)
	}

	avatarsUUID.Dimensions = append(avatarsUUID.Dimensions,
		pgtype.ArrayDimension{
			Length:     int32(len(avatarsUUID.Elements)),
			LowerBound: 1,
		})

	return &user{
		ID:       id,
		Email:    u.Email,
		Name:     u.Name,
		PassHash: passHash,
		Avatars:  avatarsUUID,
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
	avatars := make([]uuid.UUID, len(u.Avatars.Elements))
	for i := range u.Avatars.Elements {
		avatars[i] = uuid.Must(uuid.FromBytes(u.Avatars.Elements[i].Bytes[:]))
	}

	return &app.User{
		ID:        u.ID.Bytes,
		Email:     u.Email,
		Name:      u.Name,
		PassHash:  u.PassHash.Bytes,
		Avatars:   avatars,
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

		err := db.GetContext(ctx, &id, query, newUser.Email, newUser.Name, newUser.PassHash)
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
    		pass_hash = $3,
		    avatars   = $4
		where id = $5`

		_, err := db.ExecContext(ctx, query, updateUser.Email, updateUser.Name, updateUser.PassHash, updateUser.Avatars, updateUser.ID)
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
