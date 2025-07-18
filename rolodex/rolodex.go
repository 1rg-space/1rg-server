package rolodex

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"1rg-server/config"
	"1rg-server/templates"

	"github.com/gorilla/csrf"
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
	Fedi         string // "@foo@cosocial.ca"
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
	templates.RenderTemplate(w, "rolodex_add", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

// AddPostHandler handle adding a *new* user
func (h *Handler) AddPostHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: does the CSRF lib mean this max limit is useless?
	// Because it already parses the form internally?
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

func (h *Handler) EditGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var msg string
	if r.URL.Query().Get("msg") == "1" {
		msg = "Password incorrect"
	}

	row := h.db.QueryRow(`
		SELECT
		name, last_name, pronouns, email, bio, birthday, website, bluesky,
		goodreads, fedi, github, instagram, signal, phone
		FROM rolodex WHERE id=?
		`, id)
	var u user
	err = row.Scan(&u.Name, &u.LastName, &u.Pronouns, &u.Email, &u.Bio,
		&u.Birthday, &u.Website, &u.Bluesky, &u.Goodreads, &u.Fedi, &u.GitHub, &u.Instagram,
		&u.Signal, &u.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templates.RenderTemplate(w, "rolodex_edit", map[string]interface{}{
		"msg":            msg,
		"user":           u,
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func (h *Handler) EditPostHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: does the CSRF lib mean this max limit is useless?
	// Because it already parses the form internally?
	err := r.ParseMultipartForm(maxImageSize + 1<<20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process:
	// Compare against password hash in db
	// Do UPDATE with all values from form
	// Replace avatar only if a new one is provided
	// Redirect to rolodex

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var passwordHash []byte
	row := tx.QueryRow(`SELECT password_hash FROM rolodex WHERE id=?`, id)
	if err := row.Scan(&passwordHash); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ok, err := argon2.VerifyEncoded([]byte(r.PostFormValue("password")), passwordHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		// For user convenience keep them on the edit page, just with a bad password message
		http.Redirect(w, r, "/rolodex/edit/"+r.PathValue("id")+"?msg=1", http.StatusSeeOther)
		return
	}

	// Update DB with changed fields
	fields := make([]string, 0)
	vals := make([]any, 0)
	for key, val := range r.MultipartForm.Value {
		if key == "password" || key == "avatar" || key == "gorilla.csrf.Token" {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s=?", key))
		vals = append(vals, val[0])
	}

	// Add id
	vals = append(vals, r.PathValue("id"))
	query := `UPDATE rolodex SET ` + strings.Join(fields, `, `) + ` WHERE id=?`
	log.Print(query)
	_, err = tx.Exec(query, vals...)
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

	http.Redirect(w, r, "/rolodex", http.StatusSeeOther)
}
