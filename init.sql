-- init.sql
CREATE DATABASE IF NOT EXISTS uploady;

USE uploady;

-- Example table creation
CREATE TABLE IF NOT EXISTS users (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `firstName` VARCHAR(100) NOT NULL,
    `lastName` VARCHAR(100) NOT NULL,
    `email` VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
     `createdAt` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        
         PRIMARY KEY (id),

);


CREATE TABLE IF NOT EXISTS receipts (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `userId` UNSIGNED INT UNSIGNED NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `amount` DECIMAL(10, 2) NOT NULL,
    `description` TEXT,
    `imagePath` VARCHAR(255) NOT NULL,
    `date` TIMESTAMP NOT NULL,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    FOREIGN KEY (`userId`) REFERENCES users(`id`)
);