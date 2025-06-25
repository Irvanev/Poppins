package repository

import (
	"database/sql"
	"poppins/domain"
)

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) Create(u *domain.User) error {
	return r.DB.QueryRow(
		`INSERT INTO users(username, first_name, phone) VALUES($1,$2,$3) RETURNING id,created_at`,
		u.Username, u.FirstName, u.Phone,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepo) GetByID(id int64) (*domain.User, error) {
	u := &domain.User{}
	err := r.DB.QueryRow(
		`SELECT id,username,first_name,phone,created_at FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Username, &u.FirstName, &u.Phone, &u.CreatedAt)
	return u, err
}

func (r *UserRepo) Delete(id int64) (*domain.User, error) {
	u, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	_, err = r.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	return u, err
}
