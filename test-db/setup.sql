-- setup.sql (MySQL compatible)

DROP TABLE IF EXISTS urls;

CREATE TABLE urls (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shortUrl VARCHAR(20) UNIQUE NOT NULL,
    longUrl TEXT NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    clicks INT DEFAULT 0,
    lastClicked TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_shortUrl (shortUrl),
    INDEX idx_longUrl_hash (longUrl(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add some test data
INSERT INTO urls (shortUrl, longUrl) VALUES
('test123', 'https://www.example.com'),
('demo456', 'https://www.google.com');
