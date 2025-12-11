package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/zuquanzhi/Chirp/backend/internal/config"
	"github.com/zuquanzhi/Chirp/backend/internal/domain"
	handler "github.com/zuquanzhi/Chirp/backend/internal/handler/http"
	"github.com/zuquanzhi/Chirp/backend/internal/repository/mysql"
	"github.com/zuquanzhi/Chirp/backend/internal/repository/sqlite"
	"github.com/zuquanzhi/Chirp/backend/internal/service"
	"github.com/zuquanzhi/Chirp/backend/pkg/limiter"
	"github.com/zuquanzhi/Chirp/backend/pkg/logger"
	"github.com/zuquanzhi/Chirp/backend/pkg/sms"
)

func main() {
	logFile, err := logger.Setup("logs")
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer logFile.Close()

	// Load Config
	cfg := config.Load()

	// Init Infrastructure (DB, FS)
	var db *sql.DB

	switch cfg.DBDriver {
	case "mysql":
		db, err = mysql.InitDB(cfg.DBDSN)
	case "sqlite":
		db, err = sqlite.InitDB(cfg.SQLitePath)
	default:
		log.Fatalf("unsupported DB_DRIVER: %s", cfg.DBDriver)
	}
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer db.Close()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("create uploads dir: %v", err)
	}

	// Init Repositories
	var (
		userRepo     domain.UserRepository
		codeRepo     domain.VerificationCodeRepository
		resourceRepo domain.ResourceRepository
	)

	switch cfg.DBDriver {
	case "mysql":
		userRepo = mysql.NewUserRepository(db)
		codeRepo = mysql.NewCodeRepository(db)
		resourceRepo = mysql.NewResourceRepository(db)
	case "sqlite":
		userRepo = sqlite.NewUserRepository(db)
		codeRepo = sqlite.NewCodeRepository(db)
		resourceRepo = sqlite.NewResourceRepository(db)
	default:
		log.Fatalf("unsupported DB_DRIVER: %s", cfg.DBDriver)
	}

	// Init Services
	// Rate Limiter: 1 request per minute per phone number
	rateLimiter := limiter.NewInMemoryLimiter(1, time.Minute)

	// SMS Sender: Use Aliyun if configured, else Console
	var smsSender sms.Sender
	if cfg.AliyunAccessKeyID != "" {
		smsSender = sms.NewAliyunSender(
			cfg.AliyunAccessKeyID,
			cfg.AliyunAccessKeySecret,
			cfg.AliyunSignName,
			cfg.AliyunTemplateCode,
		)
		log.Println("Using Aliyun SMS Sender")
	} else {
		smsSender = &sms.ConsoleSender{}
		log.Println("Using Console SMS Sender (Mock)")
	}

	authSvc := service.NewAuthService(userRepo, codeRepo, smsSender, rateLimiter, cfg.JWTSecret)

	// Init Storage
	var storage service.FileStorage
	var storageErr error

	switch cfg.StorageBackend {
	case "oss":
		log.Println("Using Aliyun OSS Storage")
		storage, storageErr = service.NewAliyunOSSStorage(
			cfg.AliyunEndpoint,
			cfg.AliyunAccessKeyID,
			cfg.AliyunAccessKeySecret,
			cfg.AliyunBucketName,
		)
	case "local":
		log.Println("Using Local File Storage")
		storage, storageErr = service.NewLocalStorage(cfg.UploadDir)
	default:
		log.Fatalf("unsupported STORAGE_BACKEND: %s", cfg.StorageBackend)
	}

	if storageErr != nil {
		log.Fatalf("failed to init storage: %v", storageErr)
	}
	resourceSvc := service.NewResourceService(resourceRepo, storage)

	// Init Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	resourceHandler := handler.NewResourceHandler(resourceSvc)

	// Setup Router
	r := mux.NewRouter()
	r.Use(handler.RecoverMiddleware)
	r.Use(handler.LoggingMiddleware)

	// Public Routes
	r.HandleFunc("/signup", authHandler.Signup).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Phone Auth Routes
	r.HandleFunc("/auth/send-code", authHandler.SendCode).Methods("POST")
	r.HandleFunc("/signup/phone", authHandler.SignupPhone).Methods("POST")
	r.HandleFunc("/login/phone", authHandler.LoginPhone).Methods("POST")

	publicRes := r.PathPrefix("/api/public").Subrouter()
	// Use OptionalAuthMiddleware to attach user info if token is present
	publicRes.Use(handler.OptionalAuthMiddleware(authSvc, cfg.JWTSecret))
	publicRes.HandleFunc("/resources", resourceHandler.Upload).Methods("POST")
	publicRes.HandleFunc("/resources", resourceHandler.List).Methods("GET")
	publicRes.HandleFunc("/resources/{id}/download", resourceHandler.Download).Methods("GET")

	// Protected Routes (User Profile, etc.)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(handler.AuthMiddleware(authSvc, cfg.JWTSecret))

	api.HandleFunc("/me", authHandler.Me).Methods("GET")
	api.HandleFunc("/me", authHandler.UpdateMe).Methods("PATCH")
	// api.HandleFunc("/resources", resourceHandler.Upload).Methods("POST") // Moved to public for MVP 1.0

	// Admin Routes (Review, etc.) - In real app, check for Admin role
	admin := r.PathPrefix("/api/admin").Subrouter()
	admin.Use(handler.AuthMiddleware(authSvc, cfg.JWTSecret))
	admin.Use(handler.AdminMiddleware)
	admin.HandleFunc("/resources/{id}/review", resourceHandler.Review).Methods("POST")
	admin.HandleFunc("/resources/duplicates", resourceHandler.CheckDuplicate).Methods("GET")

	// Static files (optional, usually handled by Nginx)
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Start Server
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + cfg.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Chirp server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
