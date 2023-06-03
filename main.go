package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	Username string
	Password string
}

type LogEntry struct {
	Username  string
	Status    string
	Timestamp time.Time
	Hostname  string
	IPAddress string
}

var db *sql.DB

func main() {
	// Load config.env file
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Error loading config.env file")
	}

	// PostgreSQL connection settings
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// PostgreSQL connection check
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PostgreSQL connection successful")

	// Create "users" table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username VARCHAR(50) PRIMARY KEY,
			password VARCHAR(50)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Users table created or already exists")

	// Create "logs" table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50),
			status VARCHAR(10),
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			hostname VARCHAR(100),
			ip_address VARCHAR(50)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logs table created or already exists")

	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	fmt.Println("Web server started. Visit http://localhost:9006 to access the login page.")
	log.Fatal(http.ListenAndServe(":9006", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		renderTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check user credentials in the database
	user, err := getUser(username)
	if err != nil {
		logEntry(username, "failed", r.Host, r.RemoteAddr)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if password != user.Password {
		logEntry(username, "failed", r.Host, r.RemoteAddr)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	logEntry(username, "success", r.Host, r.RemoteAddr)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	command := r.FormValue("command")
	var output string

	switch command {
	case "cd":
		output = cd()
	case "ls":
		output = ls()
	case "clear":
		output = clear()
	default:
		output = "Unknown command"
	}

	renderTemplate(w, "dashboard.html", output)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getUser(username string) (*User, error) {
	query := "SELECT * FROM users WHERE username = $1"
	row := db.QueryRow(query, username)

	var user User
	err := row.Scan(&user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func logEntry(username string, status string, hostname string, ipAddress string) {
	_, err := db.Exec(`
		INSERT INTO logs (username, status, hostname, ip_address)
		VALUES ($1, $2, $3, $4)
	`, username, status, hostname, ipAddress)
	if err != nil {
		log.Println("Failed to insert log entry:", err)
	}
}

func cd() string {
	return `unfortunately, i cannot afford more directories.
if you want to help, you can ... '.`
}

func ls() string {
	return "This is the list of files: file1.txt, file2.txt, file3.txt"
}

func clear() string {
	return ""
}
