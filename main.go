package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed frontend/dist
var staticFiles embed.FS

var (
	db       *gorm.DB
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for LAN access
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	broadcast = make(chan []byte)
)

// Database models
type Preset struct {
	ID         uint   `gorm:"primaryKey" json:"-"`
	KeyNum     string `gorm:"column:key_num" json:"key_num"`
	KeyName    string `gorm:"column:key_name" json:"key_name"`
	Status     string `gorm:"column:status" json:"status"`
	TransTime  string `gorm:"column:transition_time" json:"transition_time"`
	CurPreset  string `gorm:"column:current_preset" json:"current_preset"`
	ContentRaw string `gorm:"column:content_raw;type:text" json:"-"`
	Content    []PresetContent `gorm:"-" json:"content"`
}

type PresetContent struct {
	Num           string  `json:"num,omitempty"`
	Person        string  `json:"person,omitempty"`
	Name          string  `json:"name,omitempty"`
	SongName      string  `json:"song_name,omitempty"`
	CurrentLyrics *int    `json:"current_lyrics,omitempty"`
	Lyrics        []Lyric `json:"lyrics,omitempty"`
}

type Lyric struct {
	TransitionTime interface{} `json:"transition_time"` // Can be string or number
	Text           string      `json:"text"`
}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("live-titler.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&Preset{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize with default presets if empty
	var count int64
	db.Model(&Preset{}).Count(&count)
	if count == 0 {
		initDefaultData()
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

func initDefaultData() {
	// Default presets - simplified version
	defaultPresets := []Preset{
		{
			KeyNum:    "0",
			KeyName:   "KEY0",
			Status:    "CLOSED",
			TransTime: "1",
			CurPreset: "0",
			Content: []PresetContent{
				{Num: "0", Person: "表演者", Name: "节目名"},
			},
		},
		{
			KeyNum:    "1",
			KeyName:   "KEY1",
			Status:    "CLOSED",
			TransTime: "1",
			CurPreset: "0",
			Content: []PresetContent{
				{Num: "0", Person: "表演者", Name: "节目名"},
			},
		},
		{
			KeyNum:    "2",
			KeyName:   "KEY2",
			Status:    "CLOSED",
			TransTime: "1",
			CurPreset: "0",
			Content: []PresetContent{
				{
					SongName:      "默认歌曲",
					CurrentLyrics: intPtr(0),
					Lyrics: []Lyric{
						{TransitionTime: "1", Text: ""},
					},
				},
			},
		},
		{
			KeyNum:    "3",
			KeyName:   "KEY3",
			Status:    "CLOSED",
			TransTime: "1",
			CurPreset: "0",
			Content: []PresetContent{
				{
					SongName:      "默认歌曲",
					CurrentLyrics: intPtr(0),
					Lyrics: []Lyric{
						{TransitionTime: "1", Text: ""},
					},
				},
			},
		},
	}

	for _, preset := range defaultPresets {
		contentJSON, _ := json.Marshal(preset.Content)
		preset.ContentRaw = string(contentJSON)
		db.Create(&preset)
	}
}

// BeforeSave hook to serialize Content to ContentRaw
func (p *Preset) BeforeSave(tx *gorm.DB) error {
	if len(p.Content) > 0 {
		contentJSON, err := json.Marshal(p.Content)
		if err != nil {
			return err
		}
		p.ContentRaw = string(contentJSON)
	}
	return nil
}

// AfterFind hook to deserialize ContentRaw to Content
func (p *Preset) AfterFind(tx *gorm.DB) error {
	if p.ContentRaw != "" {
		err := json.Unmarshal([]byte(p.ContentRaw), &p.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

// HTTP Handlers
func getPresets(w http.ResponseWriter, r *http.Request) {
	var presets []Preset
	result := db.Find(&presets)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presets)
}

func updatePreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyNum := vars["keyNum"]

	var preset Preset
	if err := json.NewDecoder(r.Body).Decode(&preset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Serialize Content to ContentRaw before updating
	if len(preset.Content) > 0 {
		contentJSON, err := json.Marshal(preset.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		preset.ContentRaw = string(contentJSON)
	}

	// Update the preset in database
	result := db.Model(&Preset{}).Where("key_num = ?", keyNum).Updates(map[string]interface{}{
		"key_name":        preset.KeyName,
		"status":          preset.Status,
		"transition_time": preset.TransTime,
		"current_preset":  preset.CurPreset,
		"content_raw":     preset.ContentRaw,
	})

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the update to all connected clients
	var allPresets []Preset
	db.Find(&allPresets)
	presetsJSON, _ := json.Marshal(allPresets)
	broadcast <- presetsJSON

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func importPresets(w http.ResponseWriter, r *http.Request) {
	var presets []Preset
	if err := json.NewDecoder(r.Body).Decode(&presets); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Clear existing presets
	db.Exec("DELETE FROM presets")

	// Insert new presets
	for _, preset := range presets {
		db.Create(&preset)
	}

	// Broadcast the update
	presetsJSON, _ := json.Marshal(presets)
	broadcast <- presetsJSON

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func exportPresets(w http.ResponseWriter, r *http.Request) {
	var presets []Preset
	db.Find(&presets)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presets)
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// Send current state to new client
	var presets []Preset
	db.Find(&presets)
	presetsJSON, _ := json.Marshal(presets)
	conn.WriteMessage(websocket.TextMessage, presetsJSON)

	// Keep connection alive and listen for client disconnect
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			break
		}
	}
}

// Broadcast messages to all connected clients
func handleBroadcast() {
	for {
		message := <-broadcast
		clientsMu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
}

func main() {
	initDB()

	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/presets", getPresets).Methods("GET")
	router.HandleFunc("/api/presets/{keyNum}", updatePreset).Methods("PUT")
	router.HandleFunc("/api/presets/import", importPresets).Methods("POST")
	router.HandleFunc("/api/presets/export", exportPresets).Methods("GET")
	router.HandleFunc("/ws", handleWebSocket)

	// Serve static files from frontend/dist
	subFS, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	
	// Serve index.html for all non-API routes (SPA routing)
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		
		// Try to open the file
		file, err := subFS.Open(path[1:]) // Remove leading slash
		if err != nil {
			// If file doesn't exist, serve index.html for SPA routing
			indexFile, err := subFS.Open("index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()
			
			stat, _ := indexFile.(interface{ Stat() (fs.FileInfo, error) }).Stat()
			http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(interface{ Read([]byte) (int, error); Seek(int64, int) (int64, error) }))
			return
		}
		defer file.Close()
		
		// Serve the file
		http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
	}))

	// Start broadcast handler
	go handleBroadcast()

	fmt.Println("Server starting on :3001")
	fmt.Println("Control Panel: http://localhost:3001/control-panel")
	fmt.Println("Display: http://localhost:3001/show-source")
	
	log.Fatal(http.ListenAndServe(":3001", router))
}
