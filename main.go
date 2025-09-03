package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/dlclark/regexp2"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

//go:embed templates/*.html
var tmplFS embed.FS

var (
	formTpl    = template.Must(template.ParseFS(tmplFS, "templates/form.html"))
	messageTpl = template.Must(template.ParseFS(tmplFS, "templates/message.html"))
)

var (
	dbHost           = os.Getenv("DATABASE_HOST")
	dbPort           = os.Getenv("DATABASE_PORT")
	dbName           = os.Getenv("DATABASE_NAME")
	appEnv           = os.Getenv("APP_ENV")
	newPassRegex     = regexp2.MustCompile(os.Getenv("NEW_PASSWORD_REGEX"), 0)
	newPassRegexDesc = os.Getenv("NEW_PASSWORD_REGEX_DESCRIPTION")
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		formTpl.Execute(w, map[string]string{
			"Env": appEnv,
		})
	})

	http.HandleFunc("/change", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		newPassword := r.FormValue("new_password")

		if newPassword == password {
			messageTpl.Execute(w, map[string]string{
				"Error": "New password cannot be the same as the current password.",
			})
			return
		}

		if m, _ := newPassRegex.MatchString(newPassword); !m {
			messageTpl.Execute(w, map[string]string{
				"Error": "New password does not meet the requirements: " + newPassRegexDesc,
			})
			return
		}

		connStr := fmt.Sprintf("host=%s port=%s  dbname=%s  user=%s password=%s sslmode=disable", dbHost, dbPort, dbName, username, password)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			slog.Default().Error("DB connection failed", "username", username, "error", err)
			messageTpl.Execute(w, map[string]string{
				"Error": "DB connection failed: " + err.Error(),
			})
			return
		}
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf(`ALTER USER %s WITH PASSWORD %s`, pq.QuoteIdentifier(username), pq.QuoteLiteral(newPassword)))
		if err != nil {
			slog.Default().Error("Password change failed", "username", username, "error", err)
			messageTpl.Execute(w, map[string]string{
				"Error": "Password change failed: " + err.Error(),
			})
			return
		}

		slog.Default().Info("Password changed successfully", "username", username)
		messageTpl.Execute(w, map[string]string{
			"Message": "Password changed successfully!",
		})
	})

	slog.Default().Info("Server started at :8080")
	slog.Default().Error("Server error", "error", http.ListenAndServe(":8080", nil))
}
