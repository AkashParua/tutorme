CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS questions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  user_id INTEGER REFERENCES users(id)
);

CREATE  TABLE IF NOT EXISTS answers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT NOT NULL,
  created_at TIMESTAMP CURRENT_TIMESTAMP,
  question_id INTEGER REFERENCES questions(id),
  user_id INTEGER REFERENCES users(id),
  book_name VARCHAR(50),
  file_link VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS books (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_name VARCHAR(50),
  added_on TIMESTAMP CURRENT_TIMESTAMP,
  file_link VARCHAR(255),
  user_id INTEGER REFERENCES users(id)
);

INSERT INTO users (username, password, email) VALUES ('john_doe', crypt('password123', gen_salt('bf')), 'john_doe@example.com');

DELETE FROM users WHERE username = 'Akash Parua';