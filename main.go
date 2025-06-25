package main

import (
	"context"
	"database/sql"
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
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment")
	}

	cfg := config.LoadConfig()

	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatal(err)
		}
	}

	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	migrations, _ := os.ReadFile("schema.sql")
	if _, err := db.Exec(string(migrations)); err != nil {
		log.Fatal("migrations failed:", err)
	}

	userRepo := repository.NewUserRepo(db)
	adRepo := repository.NewAdRepo(db)
	uh := handlers.NewUserHandler(userRepo)
	ah := handlers.NewAdHandler(adRepo, minioClient, cfg.MinIOBucket)

	r := router.NewRouter(uh, ah)
	log.Println("Server started on port", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}
