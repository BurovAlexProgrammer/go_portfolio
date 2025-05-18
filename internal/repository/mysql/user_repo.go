package mysql

import (
	"GoPortfolio/internal/model"
	"context"
	"database/sql"
)

type MysqlUserRepo struct {
	db *sql.DB
}

func NewMysqlUserRepo(db *sql.DB) *MysqlUserRepo {
	return &MysqlUserRepo{db: db}
}

func (repo *MysqlUserRepo) Create(ctx context.Context, user *model.User) error {
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

func (repo *MysqlUserRepo) List(ctx context.Context) ([]*model.User, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, name, telegram FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.Id, &user.Name, &user.Telegram)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
