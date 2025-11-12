package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// DTOs for API responses
type KeyResponse struct {
	ID              uint              `json:"id"`
	Name            string            `json:"name"`
	Position        int               `json:"position"`
	KeyType         string            `json:"key_type"`
	Status          string            `json:"status"`
	TransitionTime  float64           `json:"transition_time"`
	CurrentPresetID *uint             `json:"current_preset_id"`
	Version         int               `json:"version"`
	Programs        []ProgramResponse `json:"programs,omitempty"`
	Songs           []SongResponse    `json:"songs,omitempty"`
}

type ProgramResponse struct {
	ID       uint   `json:"id"`
	Num      string `json:"num"`
	Name     string `json:"name"`
	Person   string `json:"person"`
	Position int    `json:"position"`
}

type SongResponse struct {
	ID             uint            `json:"id"`
	Name           string          `json:"name"`
	CurrentLyricID *uint           `json:"current_lyric_id"`
	Position       int             `json:"position"`
	Lyrics         []LyricResponse `json:"lyrics,omitempty"`
}

type LyricResponse struct {
	ID             uint    `json:"id"`
	Text           string  `json:"text"`
	TransitionTime float64 `json:"transition_time"`
	Position       int     `json:"position"`
}

// Get all keys with their associated data
func getKeys(w http.ResponseWriter, r *http.Request) {
	var keys []Key
	if err := db.Order("position").Find(&keys).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []KeyResponse
	for _, key := range keys {
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

		response = append(response, keyResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get single key
func getKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var key Key
	if err := db.First(&key, id).Error; err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keyResp)
}

// Create new key
func createKey(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name           string  `json:"name"`
		KeyType        string  `json:"key_type"` // "program" or "lyrics"
		Position       *int    `json:"position"` // Optional, will append if not provided
		TransitionTime float64 `json:"transition_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate key type
	if input.KeyType != "program" && input.KeyType != "lyrics" {
		http.Error(w, "Invalid key_type, must be 'program' or 'lyrics'", http.StatusBadRequest)
		return
	}

	// Determine position
	position := 0
	if input.Position != nil {
		position = *input.Position
		// Shift other keys
		db.Model(&Key{}).Where("position >= ?", position).Update("position", gorm.Expr("position + 1"))
	} else {
		// Get max position
		db.Model(&Key{}).Select("COALESCE(MAX(position), -1) + 1").Scan(&position)
	}

	key := Key{
		Name:           input.Name,
		Position:       position,
		KeyType:        input.KeyType,
		Status:         "CLOSED",
		TransitionTime: input.TransitionTime,
	}

	if input.TransitionTime == 0 {
		key.TransitionTime = 1.0
	}

	if err := db.Create(&key).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key)
}

// Update key
func updateKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name            *string  `json:"name"`
		Status          *string  `json:"status"`
		TransitionTime  *float64 `json:"transition_time"`
		CurrentPresetID *uint    `json:"current_preset_id"`
		Version         int      `json:"version"` // For optimistic locking
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var key Key
	if err := db.First(&key, id).Error; err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	// Optimistic locking check
	if key.Version != input.Version {
		http.Error(w, "Version conflict - key was modified by another client", http.StatusConflict)
		return
	}

	// Update fields
	// Special rule: ignore PLAYING_FORWARD if current status is not OPENED
	ignorePlayingForward := false
	if input.Status != nil && *input.Status == "PLAYING_FORWARD" && key.Status != "OPENED" {
		ignorePlayingForward = true
	}
	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Status != nil && !ignorePlayingForward {
		updates["status"] = *input.Status
	}
	if input.TransitionTime != nil {
		updates["transition_time"] = *input.TransitionTime
	}
	if input.CurrentPresetID != nil {
		updates["current_preset_id"] = *input.CurrentPresetID
	}
	updates["version"] = key.Version + 1

	if err := db.Model(&key).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle server-side state transitions (skip if ignored)
	if input.Status != nil && !ignorePlayingForward {
		handleStateTransition(uint(id), *input.Status, key.TransitionTime, key.KeyType)
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "version": key.Version + 1})
}

// Delete key
func deleteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var key Key
	if err := db.First(&key, id).Error; err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	// Delete key (cascade will handle related data)
	if err := db.Delete(&key).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reorder remaining keys
	db.Model(&Key{}).Where("position > ?", key.Position).Update("position", gorm.Expr("position - 1"))

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Reorder keys
func reorderKeys(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []uint `json:"ids"` // New order of key IDs
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update positions in transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		// First set all positions to negative values to avoid conflicts
		for i, keyID := range input.IDs {
			if err := tx.Model(&Key{}).Where("id = ?", keyID).Update("position", -(i + 1)).Error; err != nil {
				return err
			}
		}
		// Then set them to the correct positive values
		for i, keyID := range input.IDs {
			if err := tx.Model(&Key{}).Where("id = ?", keyID).Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Broadcast updates to all connected clients
func broadcastUpdate() {
	var keys []Key
	db.Order("position").Find(&keys)

	var response []KeyResponse
	for _, key := range keys {
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

		response = append(response, keyResp)
	}

	data, _ := json.Marshal(response)
	broadcast <- data
}

// Program handlers
func createProgram(w http.ResponseWriter, r *http.Request) {
	var input struct {
		KeyID    uint   `json:"key_id"`
		Num      string `json:"num"`
		Name     string `json:"name"`
		Person   string `json:"person"`
		Position *int   `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	position := 0
	if input.Position != nil {
		position = *input.Position
		db.Model(&Program{}).Where("key_id = ? AND position >= ?", input.KeyID, position).Update("position", gorm.Expr("position + 1"))
	} else {
		db.Model(&Program{}).Where("key_id = ?", input.KeyID).Select("COALESCE(MAX(position), -1) + 1").Scan(&position)
	}

	program := Program{
		KeyID:    input.KeyID,
		Num:      input.Num,
		Name:     input.Name,
		Person:   input.Person,
		Position: position,
	}

	if err := db.Create(&program).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(program)
}

func updateProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Num    *string `json:"num"`
		Name   *string `json:"name"`
		Person *string `json:"person"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	if input.Num != nil {
		updates["num"] = *input.Num
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Person != nil {
		updates["person"] = *input.Person
	}

	if err := db.Model(&Program{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	var program Program
	if err := db.First(&program, id).Error; err != nil {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&program).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.Model(&Program{}).Where("key_id = ? AND position > ?", program.KeyID, program.Position).Update("position", gorm.Expr("position - 1"))

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Song handlers
func createSong(w http.ResponseWriter, r *http.Request) {
	var input struct {
		KeyID    uint   `json:"key_id"`
		Name     string `json:"name"`
		Position *int   `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	position := 0
	if input.Position != nil {
		position = *input.Position
		db.Model(&Song{}).Where("key_id = ? AND position >= ?", input.KeyID, position).Update("position", gorm.Expr("position + 1"))
	} else {
		db.Model(&Song{}).Where("key_id = ?", input.KeyID).Select("COALESCE(MAX(position), -1) + 1").Scan(&position)
	}

	song := Song{
		KeyID:    input.KeyID,
		Name:     input.Name,
		Position: position,
	}

	if err := db.Create(&song).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(song)
}

func updateSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name           *string `json:"name"`
		CurrentLyricID *uint   `json:"current_lyric_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.CurrentLyricID != nil {
		updates["current_lyric_id"] = *input.CurrentLyricID
	}

	if err := db.Model(&Song{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	var song Song
	if err := db.First(&song, id).Error; err != nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&song).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.Model(&Song{}).Where("key_id = ? AND position > ?", song.KeyID, song.Position).Update("position", gorm.Expr("position - 1"))

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Lyric handlers
func createLyric(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SongID         uint    `json:"song_id"`
		Text           string  `json:"text"`
		TransitionTime float64 `json:"transition_time"`
		Position       *int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	position := 0
	if input.Position != nil {
		position = *input.Position
		db.Model(&Lyric{}).Where("song_id = ? AND position >= ?", input.SongID, position).Update("position", gorm.Expr("position + 1"))
	} else {
		db.Model(&Lyric{}).Where("song_id = ?", input.SongID).Select("COALESCE(MAX(position), -1) + 1").Scan(&position)
	}

	if input.TransitionTime == 0 {
		input.TransitionTime = 1.0
	}

	lyric := Lyric{
		SongID:         input.SongID,
		Text:           input.Text,
		TransitionTime: input.TransitionTime,
		Position:       position,
	}

	if err := db.Create(&lyric).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lyric)
}

func updateLyric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid lyric ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Text           *string  `json:"text"`
		TransitionTime *float64 `json:"transition_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	if input.Text != nil {
		updates["text"] = *input.Text
	}
	if input.TransitionTime != nil {
		updates["transition_time"] = *input.TransitionTime
	}

	if err := db.Model(&Lyric{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteLyric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid lyric ID", http.StatusBadRequest)
		return
	}

	var lyric Lyric
	if err := db.First(&lyric, id).Error; err != nil {
		http.Error(w, "Lyric not found", http.StatusNotFound)
		return
	}

	if err := db.Delete(&lyric).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.Model(&Lyric{}).Where("song_id = ? AND position > ?", lyric.SongID, lyric.Position).Update("position", gorm.Expr("position - 1"))

	broadcastUpdate()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
