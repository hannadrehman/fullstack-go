package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps the sql.DB connection with retry-on-startup.
type DB struct {
	*sql.DB
}

func NewDB(url string) (*DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", url)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	avatar TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS scraps (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS replies (
	id SERIAL PRIMARY KEY,
	scrap_id INTEGER NOT NULL REFERENCES scraps(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Wall scraps: a scrap written on someone's scrapbook. NULL = global feed post.
-- ALTER is needed because CREATE TABLE IF NOT EXISTS skips existing databases.
ALTER TABLE scraps ADD COLUMN IF NOT EXISTS to_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
`
	_, err := db.Exec(schema)
	return err
}

// ---- models ----

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Scrap struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Author     string    `json:"author"`
	ToUserID   *int64    `json:"to_user_id"`
	Recipient  string    `json:"recipient,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	ReplyCount int       `json:"reply_count"`
}

type Reply struct {
	ID        int64     `json:"id"`
	ScrapID   int64     `json:"scrap_id"`
	UserID    int64     `json:"user_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ---- queries ----

func (d *DB) CreateUser(ctx context.Context, name, email, avatar string) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || !strings.Contains(email, "@") {
		return nil, ErrValidation
	}
	u := &User{}
	err := d.QueryRowContext(ctx,
		`INSERT INTO users (name, email, avatar) VALUES ($1,$2,$3)
		 ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id, name, email, avatar, created_at`,
		name, email, avatar,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Avatar, &u.CreatedAt)
	return u, err
}

func (d *DB) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := d.QueryContext(ctx, `SELECT id, name, email, avatar, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Avatar, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
// note: midpoint of this file

func (d *DB) GetUser(ctx context.Context, id int64) (*User, error) {
	u := &User{}
	err := d.QueryRowContext(ctx,
		`SELECT id, name, email, avatar, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Avatar, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// scrapSelect is the shared projection for every scrap read. Keep it in sync
// with scanScrap.
const scrapSelect = `
	SELECT s.id, s.user_id, u.name, s.to_user_id, COALESCE(ru.name, ''), s.content, s.created_at,
	       (SELECT count(*) FROM replies r WHERE r.scrap_id = s.id)
	FROM scraps s
	JOIN users u ON u.id = s.user_id
	LEFT JOIN users ru ON ru.id = s.to_user_id`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanScrap(row rowScanner) (Scrap, error) {
	var s Scrap
	var toUserID sql.NullInt64
	err := row.Scan(&s.ID, &s.UserID, &s.Author, &toUserID, &s.Recipient, &s.Content, &s.CreatedAt, &s.ReplyCount)
	if toUserID.Valid {
		s.ToUserID = &toUserID.Int64
	}
	return s, err
}

func scanScraps(rows *sql.Rows) ([]Scrap, error) {
	defer rows.Close()
	out := []Scrap{}
	for rows.Next() {
		s, err := scanScrap(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CreateScrap posts a scrap. toUserID nil means a global feed post; otherwise
// the scrap lands on that member's scrapbook.
func (d *DB) CreateScrap(ctx context.Context, userID int64, toUserID *int64, content string) (*Scrap, error) {
	content = strings.TrimSpace(content)
	if content == "" || userID <= 0 || (toUserID != nil && *toUserID <= 0) {
		return nil, ErrValidation
	}
	var to sql.NullInt64
	if toUserID != nil {
		if _, err := d.GetUser(ctx, *toUserID); errors.Is(err, ErrNotFound) {
			return nil, ErrValidation
		} else if err != nil {
			return nil, err
		}
		to = sql.NullInt64{Int64: *toUserID, Valid: true}
	}
	var id int64
	err := d.QueryRowContext(ctx,
		`INSERT INTO scraps (user_id, to_user_id, content) VALUES ($1,$2,$3) RETURNING id`,
		userID, to, content,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return d.GetScrap(ctx, id)
}

func (d *DB) ListScraps(ctx context.Context, limit int) ([]Scrap, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := d.QueryContext(ctx, scrapSelect+` WHERE s.to_user_id IS NULL ORDER BY s.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return scanScraps(rows)
}

// ListUserScraps returns scraps the member wrote, wherever they were posted.
func (d *DB) ListUserScraps(ctx context.Context, userID int64, limit int) ([]Scrap, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := d.QueryContext(ctx, scrapSelect+` WHERE s.user_id = $1 ORDER BY s.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	return scanScraps(rows)
}

// ListScrapbook returns scraps written on the member's scrapbook.
func (d *DB) ListScrapbook(ctx context.Context, userID int64, limit int) ([]Scrap, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := d.QueryContext(ctx, scrapSelect+` WHERE s.to_user_id = $1 ORDER BY s.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	return scanScraps(rows)
}

func (d *DB) GetScrap(ctx context.Context, id int64) (*Scrap, error) {
	s, err := scanScrap(d.QueryRowContext(ctx, scrapSelect+` WHERE s.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (d *DB) CreateReply(ctx context.Context, scrapID, userID int64, content string) (*Reply, error) {
	content = strings.TrimSpace(content)
	if content == "" || scrapID <= 0 || userID <= 0 {
		return nil, ErrValidation
	}
	var exists bool
	if err := d.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM scraps WHERE id=$1)`, scrapID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	r := &Reply{}
	err := d.QueryRowContext(ctx,
		`INSERT INTO replies (scrap_id, user_id, content) VALUES ($1,$2,$3)
		 RETURNING id, scrap_id, user_id, content, created_at`,
		scrapID, userID, content,
	).Scan(&r.ID, &r.ScrapID, &r.UserID, &r.Content, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := d.QueryRowContext(ctx, `SELECT name FROM users WHERE id=$1`, userID).Scan(&r.Author); err != nil {
		return nil, err
	}
	return r, nil
}

func (d *DB) ListReplies(ctx context.Context, scrapID int64) ([]Reply, error) {
	rows, err := d.QueryContext(ctx, `
		SELECT r.id, r.scrap_id, r.user_id, u.name, r.content, r.created_at
		FROM replies r JOIN users u ON u.id = r.user_id
		WHERE r.scrap_id = $1 ORDER BY r.created_at ASC`, scrapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Reply
	for rows.Next() {
		var r Reply
		if err := rows.Scan(&r.ID, &r.ScrapID, &r.UserID, &r.Author, &r.Content, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
