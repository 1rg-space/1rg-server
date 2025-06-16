package rolodex

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"1rg-server/config"
	"1rg-server/templates"

	"github.com/matthewhartstonge/argon2"
)

const (
	maxImageSize  = 8 << 20 // 8 MiB
	avatarDirName = "avatars"
)

var argon = argon2.RecommendedDefaults()

// user stores user rolodex data as stored in the DB.
// Just like in the DB, there are no nulls, only empty strings.
type user struct {
	ID           int
	PasswordHash string
	Name         string
	LastName     string
	Pronouns     string // "she/her"
	Email        string
	Bio          string
	Birthday     string // "YYYY-MM-DD"
	Website      string
	Bluesky      string // "foo.bsky.social"
	Goodreads    string // "https://www.goodreads.com/user/show/<numbers>-<name>"
	Fedi         string // "https://cosocial.ca/@foo"
	GitHub       string // "username"
	Instagram    string // "username"
	Signal       string // "username"
	Phone        string // "647-555-1234"
}

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) (*Handler, error) {
	err := os.MkdirAll(filepath.Join(config.Config.AssetStorage, avatarDirName), 0755)
	if err != nil {
		return nil, err
	}
	return &Handler{db: db}, nil
}

func (h *Handler) AddGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.RenderTemplate(w, "rolodex_add", nil)
}

// AddPostHandler handle adding a *new* user
func (h *Handler) AddPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxImageSize + 1<<20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	passwordHash, err := argon.HashEncoded([]byte(r.PostFormValue("password")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store user profile and get their ID
	var id int
	err = tx.QueryRow(`
		INSERT INTO rolodex
		(password_hash, name, last_name, pronouns, email, bio, birthday, website, bluesky, goodreads, fedi,
		github, instagram, signal, phone)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		RETURNING id
		`,
		passwordHash,
		r.PostFormValue("name"), r.PostFormValue("last_name"), r.PostFormValue("pronouns"), r.PostFormValue("email"),
		r.PostFormValue("bio"), r.PostFormValue("birthday"), r.PostFormValue("website"),
		r.PostFormValue("bluesky"), r.PostFormValue("goodreads"), r.PostFormValue("fedi"),
		r.PostFormValue("github"), r.PostFormValue("instagram"), r.PostFormValue("signal"),
		r.PostFormValue("phone"),
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store their avatar
	// TODO: resize/convert image, validate the bytes
	file, _, err := r.FormFile("avatar")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err == nil {
		// A file was provided
		defer file.Close()
		f, err := os.OpenFile(
			filepath.Join(config.Config.AssetStorage, avatarDirName, strconv.Itoa(id)),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0644,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		if _, err := io.Copy(f, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect user to rolodex page where their profile will show up
	http.Redirect(w, r, "/rolodex", http.StatusSeeOther)
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT
		id, name, last_name, pronouns, email, bio, birthday, website, bluesky,
		goodreads, fedi, github, instagram, signal, phone
		FROM rolodex`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := make([]*user, 0)
	for rows.Next() {
		var u user
		err = rows.Scan(&u.ID, &u.Name, &u.LastName, &u.Pronouns, &u.Email, &u.Bio,
			&u.Birthday, &u.Website, &u.Bluesky, &u.Goodreads, &u.Fedi, &u.GitHub, &u.Instagram,
			&u.Signal, &u.Phone)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	templates.RenderTemplate(w, "rolodex", users)
}
