package rolodex

import (
	"net/http"

	"github.com/makew0rld/1rg-server/templates"
)

// user stores user rolodex data as stored in the DB.
// Just like in the DB, there are no nulls, only empty strings.
type user struct {
	ID        int
	Name      string
	Pronouns  string // "she/her"
	Email     string
	Bio       string
	Birthday  string // "MMDD"
	Website   string
	Bluesky   string // "foo.bsky.social"
	Goodreads string // "https://www.goodreads.com/user/show/<numbers>-<name>"
	Fedi      string // "https://cosocial.ca/@foo"
	GitHub    string // "username"
	Instagram string // "username"
	Signal    string // "username"
	Phone     string // "647-555-1234"
}

func AddGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.RenderTemplate(w, "rolodex_add", nil)
}

func AddPostHandler(w http.Response, r *http.Request) {
	// TODO: read POSTed data
	// - Set ID for user (how?)
	// - Resize image and store on filesystem under ID
	// - Store other data in DB
	// - Redirect user to /rolodex or show error page
}
