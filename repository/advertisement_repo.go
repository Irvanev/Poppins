package repository

import (
	"database/sql"
	"fmt"
	"poppins/domain"
	"time"
)

type AdRepo struct {
	DB *sql.DB
}

func NewAdRepo(db *sql.DB) *AdRepo {
	return &AdRepo{DB: db}
}

func (r *AdRepo) Create(ad *domain.Advertisement) error {
	var userID int64
	if err := r.DB.QueryRow(
		`SELECT id FROM users WHERE telegram_id = $1`, ad.TelegramID,
	).Scan(&userID); err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	ad.UserID = userID

	now := time.Now()
	ad.CreatedAt = now
	ad.UpdatedAt = now
	ad.Archived = false

	return r.DB.QueryRow(
		`INSERT INTO advertisements
           (user_id, title, description, price, photos_urls, address, archived, created_at, updated_at)
         VALUES
           ($1,      $2,    $3,          $4,    $5,          $6,      $7,       $8,         $9)
         RETURNING id`,
		ad.UserID,
		ad.Title,
		ad.Description,
		ad.Price,
		ad.PhotosUrls,
		ad.Address,
		ad.Archived,
		ad.CreatedAt,
		ad.UpdatedAt,
	).Scan(&ad.ID)
}

// GetByTelegramID возвращает список объявлений для пользователя с данным telegram_id.
func (r *AdRepo) GetByTelegramID(telegramID int64) ([]*domain.Advertisement, error) {
	// 1) джоин с таблицей users по telegram_id
	rows, err := r.DB.Query(
		`SELECT 
            a.id,
            a.user_id,
            u.telegram_id,
            a.title,
            a.description,
            a.price,
            a.photos_urls,
            a.address,
            a.archived,
            a.created_at,
            a.updated_at
         FROM advertisements a
         JOIN users u ON a.user_id = u.id
         WHERE u.telegram_id = $1
           AND a.archived = FALSE`,
		telegramID,
	)
	if err != nil {
		return nil, fmt.Errorf("query ads by telegram_id: %w", err)
	}
	defer rows.Close()

	var ads []*domain.Advertisement
	for rows.Next() {
		ad := &domain.Advertisement{}
		// сканируем telegram_id из таблицы users
		if err := rows.Scan(
			&ad.ID,
			&ad.UserID,
			&ad.TelegramID,
			&ad.Title,
			&ad.Description,
			&ad.Price,
			&ad.PhotosUrls,
			&ad.Address,
			&ad.Archived,
			&ad.CreatedAt,
			&ad.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad row: %w", err)
		}
		ads = append(ads, ad)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ad rows: %w", err)
	}
	return ads, nil
}

func (r *AdRepo) GetByIDAndTelegram(adID, telegramID int64) (*domain.Advertisement, error) {
	ad := &domain.Advertisement{}
	err := r.DB.QueryRow(
		`SELECT 
            a.id,
            a.user_id,
            u.telegram_id,
            a.title,
            a.description,
            a.price,
            a.photos_urls,
            a.address,
            a.archived,
            a.created_at,
            a.updated_at
         FROM advertisements a
         JOIN users u ON a.user_id = u.id
         WHERE a.id = $1
           AND u.telegram_id = $2
           AND a.archived = FALSE`,
		adID, telegramID,
	).Scan(
		&ad.ID,
		&ad.UserID,
		&ad.TelegramID,
		&ad.Title,
		&ad.Description,
		&ad.Price,
		&ad.PhotosUrls,
		&ad.Address,
		&ad.Archived,
		&ad.CreatedAt,
		&ad.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get ad by id & telegram: %w", err)
	}
	return ad, nil
}

func (r *AdRepo) Search(keyword string, maxPrice int64) ([]*domain.Advertisement, error) {
	query := `SELECT id,user_id,title,description,price,photos_urls,address,archived,created_at,updated_at
              FROM advertisements WHERE archived='false'`
	args := []interface{}{}
	i := 1

	if keyword != "" {
		query += fmt.Sprintf(" AND title ILIKE $%d", i)
		args = append(args, "%"+keyword+"%")
		i++
	}
	if maxPrice > 0 {
		query += fmt.Sprintf(" AND price <= $%d", i)
		args = append(args, maxPrice)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []*domain.Advertisement
	for rows.Next() {
		ad := &domain.Advertisement{}
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Description, &ad.Price,
			&ad.PhotosUrls, &ad.Address, &ad.Archived, &ad.CreatedAt, &ad.UpdatedAt); err != nil {
			return nil, err
		}
		ads = append(ads, ad)
	}
	return ads, rows.Err()
}

func (r *AdRepo) Update(ad *domain.Advertisement) error {
	ad.UpdatedAt = time.Now()
	return r.DB.QueryRow(
		`UPDATE advertisements SET title=$1, description=$2, price=$3, photos_urls=$4, address=$5, updated_at=$7
         WHERE id=$8 RETURNING user_id, created_at`,
		ad.Title, ad.Description, ad.Price, ad.PhotosUrls, ad.Address, ad.Archived, ad.UpdatedAt, ad.ID,
	).Scan(&ad.UserID, &ad.CreatedAt)
}

func (r *AdRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM advertisements WHERE id=$1`, id)
	return err
}

func (r *AdRepo) Archive(id int64) error {
	_, err := r.DB.Exec(`UPDATE advertisements SET archived=true, updated_at=$2 WHERE id=$1`, id, time.Now())
	return err
}
