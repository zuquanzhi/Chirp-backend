package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	createUsers := `CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255),
		email VARCHAR(255) UNIQUE,
		password VARCHAR(255),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		phone_number VARCHAR(50),
		school VARCHAR(255),
		student_id VARCHAR(50),
		birthdate VARCHAR(50),
		address TEXT,
		gender VARCHAR(20)
	);`

	createResources := `CREATE TABLE IF NOT EXISTS resources (
		id INT PRIMARY KEY AUTO_INCREMENT,
		owner_id INT,
		title VARCHAR(255),
		description TEXT,
		filename VARCHAR(255),
		original_name VARCHAR(255),
		size BIGINT,
		file_hash VARCHAR(255),
		status VARCHAR(50) DEFAULT 'PENDING',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		subject VARCHAR(255),
		type VARCHAR(50),
		FOREIGN KEY(owner_id) REFERENCES users(id)
	);`

	createNotifications := `CREATE TABLE IF NOT EXISTS notifications (
		id INT PRIMARY KEY AUTO_INCREMENT,
		user_id INT,
		content TEXT,
		is_read BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	createCodes := `CREATE TABLE IF NOT EXISTS verification_codes (
		id INT PRIMARY KEY AUTO_INCREMENT,
		phone_number VARCHAR(50),
		code VARCHAR(10),
		purpose VARCHAR(20),
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_phone_purpose (phone_number, purpose)
	);`

	if _, err := db.Exec(createUsers); err != nil {
		return nil, fmt.Errorf("create users table: %w", err)
	}
	if _, err := db.Exec(createResources); err != nil {
		return nil, fmt.Errorf("create resources table: %w", err)
	}
	if _, err := db.Exec(createNotifications); err != nil {
		return nil, fmt.Errorf("create notifications table: %w", err)
	}
	if _, err := db.Exec(createCodes); err != nil {
		return nil, fmt.Errorf("create verification_codes table: %w", err)
	}

	return db, nil
}
