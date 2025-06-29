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
		`INSERT INTO users(telegram_id, name, phone, preferred_contact) VALUES($1,$2,$3, $4) RETURNING id,created_at`,
		u.TelegramID, u.Name, u.Phone, u.PreferredContact,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepo) GetByID(telegramId string) (*domain.User, error) {
	u := &domain.User{}
	err := r.DB.QueryRow(`
        SELECT
            u.id,
            u.telegram_id,
            u.name,
            u.phone,
            u.preferred_contact,
            u.created_at,
            (
              SELECT COUNT(*)
              FROM advertisements a
              WHERE a.user_id = u.id
                AND a.archived = FALSE
            ) AS ads_count
        FROM users u
        WHERE u.telegram_id = $1
    `, telegramId).Scan(
		&u.ID,
		&u.TelegramID,
		&u.Name,
		&u.Phone,
		&u.PreferredContact,
		&u.CreatedAt,
		&u.AdsCount, // сюда сканим
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) Delete(telegramId string) (*domain.User, error) {
	u, err := r.GetByID(telegramId)
	if err != nil {
		return nil, err
	}
	_, err = r.DB.Exec(`DELETE FROM users WHERE telegram_id=$1`, telegramId)
	return u, err
}

// UpdateName обновляет только поле name
func (r *UserRepo) UpdateName(telegramId string, newName string) (*domain.User, error) {
	_, err := r.DB.Exec(
		`UPDATE users SET name = $1 WHERE telegram_id = $2`,
		newName, telegramId,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(telegramId)
}

// UpdatePhone обновляет только поле phone
func (r *UserRepo) UpdatePhone(telegramId string, newPhone string) (*domain.User, error) {
	_, err := r.DB.Exec(
		`UPDATE users SET phone = $1 WHERE telegram_id = $2`,
		newPhone, telegramId,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(telegramId)
}

// UpdatePreferredContact обновляет только поле preferred_contact
func (r *UserRepo) UpdatePreferredContact(telegramId string, newContact string) (*domain.User, error) {
	_, err := r.DB.Exec(
		`UPDATE users SET preferred_contact = $1 WHERE telegram_id = $2`,
		newContact, telegramId,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(telegramId)
}
