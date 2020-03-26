package function

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/openfaas/openfaas-cloud/sdk"
	"github.com/pkg/errors"
)

var db *sql.DB

// init establishes a persistent connection to the remote database
// the function will panic if it cannot establish a link and the
// container will restart / go into a crash/back-off loop
func init() {
	if _, err := os.Stat("/var/openfaas/secrets/password"); err == nil {
		password, _ := sdk.ReadSecret("password")
		user, _ := sdk.ReadSecret("username")
		host, _ := sdk.ReadSecret("host")
		dbName := os.Getenv("postgres_db")
		port := os.Getenv("postgres_port")
		sslmode := os.Getenv("postgres_sslmode")
		connStr := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbName + "?sslmode=" + sslmode
		var err error
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			panic(err.Error())
		}
		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}
	}
}

type Todo struct {
	ID            int        `json:"id"`
	Description   string     `json:"description"`
	CreatedDate   *time.Time `json:"created_date"`
	CompletedDate *time.Time `json:"completed_date,omitempty"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/create" {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		if err := insert(string(body)); err != nil {
			http.Error(w, fmt.Sprintf("unable to insert todo: %s", err.Error()), http.StatusInternalServerError)
			return
		}

	} else if r.Method == http.MethodGet && r.URL.Path == "/list" {
		todos, err := selectTodos()

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get todos: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		out, err := json.Marshal(todos)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to marshal todos: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}
}

func insert(description string) error {
	res, err := db.Query(`insert into todo (id, description, created_date) values (DEFAULT, $1, now());`,
		description)

	if err != nil {
		return err
	}

	defer res.Close()
	return nil
}

func selectTodos() ([]Todo, error) {
	rows, getErr := db.Query(`select id, description, created_date, completed_date from todo;`)

	if getErr != nil {
		return []Todo{}, errors.Wrap(getErr, "unable to get from todo table")
	}

	todos := []Todo{}
	defer rows.Close()
	for rows.Next() {
		result := Todo{}
		scanErr := rows.Scan(&result.ID, &result.Description, &result.CreatedDate, &result.CompletedDate)
		if scanErr != nil {
			log.Println("scan err:", scanErr)
		}
		todos = append(todos, result)
	}

	return todos, nil
}
