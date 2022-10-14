package Database

import logger "Server/Logger"

func (db *DBWrapper) ApplyMigrations() {
	db.initialMigration()
}

func (db *DBWrapper) initialMigration() {
	tx := db.Db.Exec(
		`
		CREATE DATABASE IF NOT EXISTS instabug;
		`)
	if tx.Error != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to run migration1 with this error %v", tx.Error)
	}

	tx = db.Db.Exec(
		`
		CREATE TABLE IF NOT EXISTS instabug.applications (
		id int NOT NULL AUTO_INCREMENT,
		name varchar(200) NOT NULL,
		token varchar(200) NOT NULL,
		chats_count int NOT NULL DEFAULT '0',
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY index_applications_on_token (token)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
		`)
	if tx.Error != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to run migration1 with this error %v", tx.Error)
	}

	tx = db.Db.Exec(
		`
		CREATE TABLE IF NOT EXISTS instabug.chats (
		id int NOT NULL AUTO_INCREMENT,
		application_token varchar(200) NOT NULL,
		number int NOT NULL,
		messages_count int NOT NULL DEFAULT '0',
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		PRIMARY KEY (id),
		KEY index_chats_on_application_token_and_number_and_messages_count (application_token,number,messages_count),
		CONSTRAINT fk_application_token FOREIGN KEY (application_token) REFERENCES instabug.applications (token)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
		`)
	if tx.Error != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to run migration1 with this error %v", tx.Error)
	}

	tx = db.Db.Exec(
		`
		CREATE TABLE IF NOT EXISTS instabug.messages (
		id int NOT NULL AUTO_INCREMENT,
		chat_id int NOT NULL,
		number int NOT NULL,
		body varchar(255) DEFAULT NULL,
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		PRIMARY KEY (id),
		KEY index_messages_on_chat_id_and_number (chat_id,number),
		CONSTRAINT fk_chat_id FOREIGN KEY (chat_id) REFERENCES instabug.chats (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
		`)
	if tx.Error != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to run migration1 with this error %v", tx.Error)
	}
}
