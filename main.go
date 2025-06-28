package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"poppins/config"
	"poppins/handlers"
	"poppins/repository"
	"poppins/router"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "poppins/docs"
)

// @title           Monolith Ads API
// @version         1.0
// @description     Простое REST API для работы с объявлениями и пользователями.
// @host      localhost:8080
// @BasePath  /

func main() {
	// Загружаем .env (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment")
	}

	// Конфиг
	cfg := config.LoadConfig()

	// Инициализация MinIO-клиента
	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Проверяем и создаём бакет, если нужно
	exists, err := minioClient.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatal(err)
		}
	}

	// Выставляем публичную read-only политику на весь бакет
	publicReadPolicy := fmt.Sprintf(`{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::%s/*"]
    }
  ]
}`, cfg.MinIOBucket)

	if err := minioClient.SetBucketPolicy(ctx, cfg.MinIOBucket, publicReadPolicy); err != nil {
		log.Fatalf("cannot set public policy on bucket %q: %v", cfg.MinIOBucket, err)
	}
	log.Printf("Bucket %q is now publicly readable", cfg.MinIOBucket)

	// Подключаемся к базе
	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Миграции
	migrations, _ := os.ReadFile("schema.sql")
	if _, err := db.Exec(string(migrations)); err != nil {
		log.Fatal("migrations failed:", err)
	}

	// Репозитории и хендлеры
	userRepo := repository.NewUserRepo(db)
	adRepo := repository.NewAdRepo(db)
	uh := handlers.NewUserHandler(userRepo)
	ah := handlers.NewAdHandler(adRepo, minioClient, cfg.MinIOBucket)

	// Роутер и Swagger
	r := router.NewRouter(uh, ah)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Старт сервера
	log.Println("Server started on port", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}
