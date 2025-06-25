package repository

import (
	"database/sql"
	"fmt"
	"poppins/domain"
	"time"

	"github.com/lib/pq"
)

type AdRepo struct {
	DB *sql.DB
}

func NewAdRepo(db *sql.DB) *AdRepo {
	return &AdRepo{DB: db}
}

func (r *AdRepo) Create(ad *domain.Advertisement) error {
	ad.CreatedAt = time.Now()
	ad.UpdatedAt = time.Now()
	ad.Archived = false
	return r.DB.QueryRow(
		`INSERT INTO advertisements(user_id,title,description,price,photos_urls,address,archived,created_at,updated_at)
         VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		ad.UserID, ad.Title, ad.Description, ad.Price, pq.Array(ad.PhotosUrls),
		ad.Address, ad.Archived, ad.CreatedAt, ad.UpdatedAt,
	).Scan(&ad.ID)
}

func (r *AdRepo) GetByID(id int64) (*domain.Advertisement, error) {
	ad := &domain.Advertisement{}
	err := r.DB.QueryRow(
		`SELECT id,user_id,title,description,price,photos_urls,address,archived,created_at,updated_at
         FROM advertisements WHERE id=$1 AND archived='false'`, id,
	).Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Description, &ad.Price,
		pq.Array(&ad.PhotosUrls), &ad.Address, &ad.Archived, &ad.CreatedAt, &ad.UpdatedAt)
	return ad, err
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
			pq.Array(&ad.PhotosUrls), &ad.Address, &ad.Archived, &ad.CreatedAt, &ad.UpdatedAt); err != nil {
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
		ad.Title, ad.Description, ad.Price, pq.Array(ad.PhotosUrls), ad.Address, ad.Archived, ad.UpdatedAt, ad.ID,
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
