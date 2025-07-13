package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type SensorData struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

var (
	basePath string
	db       *sql.DB
	tmpl     *template.Template
	store    = sessions.NewCookieStore([]byte("super-secret-key"))
	hub      = &WebSocketHub{clients: make(map[*websocket.Conn]bool)}
	upgrader = websocket.Upgrader{}
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	basePath = filepath.Join(filepath.Dir(filename), "../../")
}

type WebSocketHub struct {
	clients map[*websocket.Conn]bool
	lock    sync.Mutex
}

func (h *WebSocketHub) Broadcast(data SensorData) {
	h.lock.Lock()
	defer h.lock.Unlock()

	msg, _ := json.Marshal(data)
	log.Printf("[WebSocket] Broadcasting to %d clients: %s", len(h.clients), msg)

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("WebSocket write error:", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func (h *WebSocketHub) Register(conn *websocket.Conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.clients[conn] = true
}

func main() {
	tmpl = template.Must(template.ParseFiles(filepath.Join(basePath, "web/templates/login.html")))

	var err error
	db, err = sql.Open("sqlite3", filepath.Join(basePath, "data/data.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()
	createDefaultAdmin()

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", monitorHandler)
	http.HandleFunc("/sensor", sensorHandler)
	http.HandleFunc("/ws", wsHandler)

	log.Println("App Server listening on http://localhost:9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func createTable() {
	queryUser := `CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL
	)`
	querySensor := `CREATE TABLE IF NOT EXISTS sensor (
		id TEXT,
		type TEXT,
		value REAL,
		unit TEXT,
		timestamp TEXT
	)`
	if _, err := db.Exec(queryUser); err != nil {
		log.Fatal("Error creating users table:", err)
	}

	if _, err := db.Exec(querySensor); err != nil {
		log.Fatal("Error creating sensor table:", err)
	}
}

func createDefaultAdmin() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "admin").Scan(&count)
	if err != nil {
		log.Fatal("check admin user error:", err)
	}
	if count == 0 {
		_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", "admin", "admin")
		if err != nil {
			log.Fatal("insert admin error:", err)
		}
		log.Println("Default user admin:admin created")
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var dbPass string
		err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPass)
		if err != nil || password != dbPass {
			// Trả lại 1 đoạn HTML để HTMX render lỗi
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`<div class="alert alert-danger">Invalid username or password</div>`))
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["authenticated"] = true
		session.Save(r, w)

		// Nếu login thành công, trả về HTML chuyển hướng bằng HTMX
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="alert alert-success">Login successful! Redirecting...</div>
			<script>
				setTimeout(() => {
					window.location.href = "/";
				}, 1000);
			</script>`))
		return
	}

	// fallback nếu GET
	tmpl.Execute(w, nil)
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, filepath.Join(basePath, "web/templates/monitor.html"))
	// w.Write([]byte("Welcome! You are logged in."))
}

func sensorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var data SensorData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[/sensor] Received: %+v", data)

	hub.Broadcast(data)
	w.WriteHeader(http.StatusOK)

	_, err = db.Exec(`INSERT INTO sensor (id, type, value, unit, timestamp) VALUES (?, ?, ?, ?, ?)`,
		data.ID, data.Type, data.Value, data.Unit, data.Timestamp)
	if err != nil {
		http.Error(w, "DB insert error", http.StatusInternalServerError)
		return
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	log.Println("[/ws] Client connected")
	hub.Register(conn)
}
