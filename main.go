package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

//go:embed frontend/dist
var staticFiles embed.FS

var (
	db       *gorm.DB
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	broadcast = make(chan []byte)
	
	// State manager for server-side transitions
	stateManager = NewStateManager()
)

// StateTransition represents an active state transition
type StateTransition struct {
	keyID     uint
	timer     *time.Timer
	cancelCh  chan struct{}
	fromState string
	toState   string
}

// StateManager manages all active state transitions
type StateManager struct {
	mu          sync.RWMutex
	transitions map[uint]*StateTransition
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		transitions: make(map[uint]*StateTransition),
	}
}

// Cancel cancels any existing transition for the given key
func (sm *StateManager) Cancel(keyID uint) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if transition, exists := sm.transitions[keyID]; exists {
		if transition.timer != nil {
			transition.timer.Stop()
		}
		close(transition.cancelCh)
		delete(sm.transitions, keyID)
	}
}

// Start starts a new state transition
func (sm *StateManager) Start(keyID uint, duration time.Duration, fromState, toState string, callback func()) {
	sm.Cancel(keyID) // Cancel any existing transition
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	cancelCh := make(chan struct{})
	transition := &StateTransition{
		keyID:     keyID,
		cancelCh:  cancelCh,
		fromState: fromState,
		toState:   toState,
	}
	
	transition.timer = time.AfterFunc(duration, func() {
		select {
		case <-cancelCh:
			return
		default:
			callback()
			sm.mu.Lock()
			delete(sm.transitions, keyID)
			sm.mu.Unlock()
		}
	})
	
	sm.transitions[keyID] = transition
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	fmt.Println("Database initialized successfully")

	router := mux.NewRouter()

	router.HandleFunc("/api/keys", getKeys).Methods("GET")
	router.HandleFunc("/api/keys", createKey).Methods("POST")
	router.HandleFunc("/api/keys/{id}", getKey).Methods("GET")
	router.HandleFunc("/api/keys/{id}", updateKey).Methods("PATCH")
	router.HandleFunc("/api/keys/{id}", deleteKey).Methods("DELETE")
	router.HandleFunc("/api/keys/reorder", reorderKeys).Methods("POST")

	router.HandleFunc("/api/programs", createProgram).Methods("POST")
	router.HandleFunc("/api/programs/{id}", updateProgram).Methods("PATCH")
	router.HandleFunc("/api/programs/{id}", deleteProgram).Methods("DELETE")

	router.HandleFunc("/api/songs", createSong).Methods("POST")
	router.HandleFunc("/api/songs/{id}", updateSong).Methods("PATCH")
	router.HandleFunc("/api/songs/{id}", deleteSong).Methods("DELETE")

	router.HandleFunc("/api/lyrics", createLyric).Methods("POST")
	router.HandleFunc("/api/lyrics/{id}", updateLyric).Methods("PATCH")
	router.HandleFunc("/api/lyrics/{id}", deleteLyric).Methods("DELETE")

	router.HandleFunc("/ws", handleWebSocket)

	subFS, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		file, err := subFS.Open(path[1:])
		if err != nil {
			indexFile, err := subFS.Open("index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()

			stat, _ := indexFile.(interface{ Stat() (fs.FileInfo, error) }).Stat()
			http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(interface {
				Read([]byte) (int, error)
				Seek(int64, int) (int64, error)
			}))
			return
		}
		defer file.Close()

		http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
	}))

	go handleBroadcast()

	printStartupBanner()

	log.Fatal(http.ListenAndServe(":3001", router))
}

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

	sendFullState(conn)

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func sendFullState(conn *websocket.Conn) {
	var keys []Key
	db.Order("position").Find(&keys)

	var response []KeyResponse
	for _, key := range keys {
		keyResp := buildKeyResponse(key)
		response = append(response, keyResp)
	}

	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

func buildKeyResponse(key Key) KeyResponse {
	keyResp := KeyResponse{
		ID:              key.ID,
		Name:            key.Name,
		Position:        key.Position,
		KeyType:         key.KeyType,
		Status:          key.Status,
		TransitionTime:  key.TransitionTime,
		CurrentPresetID: key.CurrentPresetID,
		Version:         key.Version,
	}

	if key.KeyType == "program" {
		var programs []Program
		db.Where("key_id = ?", key.ID).Order("position").Find(&programs)
		for _, prog := range programs {
			keyResp.Programs = append(keyResp.Programs, ProgramResponse{
				ID:       prog.ID,
				Num:      prog.Num,
				Name:     prog.Name,
				Person:   prog.Person,
				Position: prog.Position,
			})
		}
	} else if key.KeyType == "lyrics" {
		var songs []Song
		db.Where("key_id = ?", key.ID).Order("position").Find(&songs)
		for _, song := range songs {
			var lyrics []Lyric
			db.Where("song_id = ?", song.ID).Order("position").Find(&lyrics)

			songResp := SongResponse{
				ID:             song.ID,
				Name:           song.Name,
				CurrentLyricID: song.CurrentLyricID,
				Position:       song.Position,
			}
			for _, lyric := range lyrics {
				songResp.Lyrics = append(songResp.Lyrics, LyricResponse{
					ID:             lyric.ID,
					Text:           lyric.Text,
					TransitionTime: lyric.TransitionTime,
					Position:       lyric.Position,
				})
			}
			keyResp.Songs = append(keyResp.Songs, songResp)
		}
	}

	return keyResp
}

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

func printStartupBanner() {
	const (
		colorReset  = "\033[0m"
		colorGreen  = "\033[32m"
		colorCyan   = "\033[36m"
		colorYellow = "\033[33m"
		colorBold   = "\033[1m"
	)

	fmt.Println()
	fmt.Printf("%s%s╔════════════════════════════════════════════════════════════╗%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s║                   Live Titler 服务已启动                   ║%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s╚════════════════════════════════════════════════════════════╝%s\n", colorBold, colorCyan, colorReset)
	fmt.Println()
	fmt.Printf("%s%s  ► 控制界面:%s http://localhost:3001/control-panel\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%s%s  ► 展示界面:%s http://localhost:3001/show-source\n", colorBold, colorYellow, colorReset)
	fmt.Println()
	fmt.Printf("%s%s提示:%s 按 Ctrl+C 停止服务\n", colorBold, colorCyan, colorReset)
	fmt.Println()
}
