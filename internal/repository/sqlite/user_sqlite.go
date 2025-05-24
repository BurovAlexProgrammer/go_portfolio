package sqlite

import (
	"GoPortfolio/internal/domain"
	"context"
	"database/sql"
)

type SqliteUserRepo struct {
	db *sql.DB
}

func NewSqliteUserRepo(db *sql.DB) *SqliteUserRepo {
	return &SqliteUserRepo{db: db}
}

func (repo *SqliteUserRepo) Create(ctx context.Context, user *domain.User) error {
	res, err := repo.db.ExecContext(ctx, "INSERT INTO users (name, telegram) VALUES (?,?)", user.Name, user.Telegram)
	if err != nil {
		return err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.Id = newId
	return nil
}

func (repo *SqliteUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, name, telegram FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(&user.Id, &user.Name, &user.Telegram)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
