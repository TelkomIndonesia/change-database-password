package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

//go:embed templates/*.html
var tmplFS embed.FS

var (
	formTpl    = template.Must(template.ParseFS(tmplFS, "templates/form.html"))
	messageTpl = template.Must(template.ParseFS(tmplFS, "templates/message.html"))
)

var (
	dbHost = os.Getenv("DATABASE_HOST")
	dbPort = os.Getenv("DATABASE_PORT")
	dbName = os.Getenv("DATABASE_NAME")
	appEnv = os.Getenv("APP_ENV")
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

		connStr := fmt.Sprintf("host=%s port=%s  dbname=%s  user=%s password=%s sslmode=disable", dbHost, dbPort, dbName, username, password)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			messageTpl.Execute(w, map[string]string{
				"Error": "DB connection failed: " + err.Error(),
			})
			return
		}
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf(`ALTER USER "%s" WITH PASSWORD '%s'`, username, newPassword))
		if err != nil {
			messageTpl.Execute(w, map[string]string{
				"Error": "Password change failed: " + err.Error(),
			})
			return
		}

		messageTpl.Execute(w, map[string]string{
			"Message": "Password changed successfully!",
		})
	})

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
