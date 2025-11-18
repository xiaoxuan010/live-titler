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
	Programs        []ProgramResponse `json:"programs"`
	Songs           []SongResponse    `json:"songs"`
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
	Lyrics         []LyricResponse `json:"lyrics"`
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
			Programs:        []ProgramResponse{},
			Songs:           []SongResponse{},
		}

		switch key.KeyType {
		case "program":
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
		case "lyrics":
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
					Lyrics:         []LyricResponse{},
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
		Programs:        []ProgramResponse{},
		Songs:           []SongResponse{},
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
				Lyrics:         []LyricResponse{},
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

	// Return a normalized KeyResponse so frontend always receives programs/songs fields
	keyResp := KeyResponse{
		ID:              key.ID,
		Name:            key.Name,
		Position:        key.Position,
		KeyType:         key.KeyType,
		Status:          key.Status,
		TransitionTime:  key.TransitionTime,
		CurrentPresetID: key.CurrentPresetID,
		Version:         key.Version,
		Programs:        []ProgramResponse{},
		Songs:           []SongResponse{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keyResp)
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
			Programs:        []ProgramResponse{},
			Songs:           []SongResponse{},
		}

		switch key.KeyType {
		case "program":
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
		case "lyrics":
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
					Lyrics:         []LyricResponse{},
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

// Reorder programs
func reorderPrograms(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []uint `json:"ids"` // New order of program IDs
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update positions in transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		for i, progID := range input.IDs {
			if err := tx.Model(&Program{}).Where("id = ?", progID).Update("position", -(i + 1)).Error; err != nil {
				return err
			}
		}
		for i, progID := range input.IDs {
			if err := tx.Model(&Program{}).Where("id = ?", progID).Update("position", i).Error; err != nil {
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

// Reorder lyrics
func reorderLyrics(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []uint `json:"ids"` // New order of lyric IDs
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update positions in transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		for i, lyricID := range input.IDs {
			if err := tx.Model(&Lyric{}).Where("id = ?", lyricID).Update("position", -(i + 1)).Error; err != nil {
				return err
			}
		}
		for i, lyricID := range input.IDs {
			if err := tx.Model(&Lyric{}).Where("id = ?", lyricID).Update("position", i).Error; err != nil {
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

// CSV Import for Programs
func importProgramsCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var input struct {
		CSV string `json:"csv"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse CSV (simple parsing, expects: num,name,person per line)
	lines := parseCSVLines(input.CSV)
	if len(lines) == 0 {
		http.Error(w, "Empty CSV", http.StatusBadRequest)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// Delete existing programs
		if err := tx.Where("key_id = ?", keyID).Delete(&Program{}).Error; err != nil {
			return err
		}

		// Create new programs
		for i, line := range lines {
			fields := parseCSVLine(line)
			if len(fields) < 3 {
				continue // Skip invalid lines
			}

			program := Program{
				KeyID:    uint(keyID),
				Num:      fields[0],
				Name:     fields[1],
				Person:   fields[2],
				Position: i,
			}

			if err := tx.Create(&program).Error; err != nil {
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

// CSV Export for Programs
func exportProgramsCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var programs []Program
	if err := db.Where("key_id = ?", keyID).Order("position").Find(&programs).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate CSV
	csv := ""
	for _, prog := range programs {
		csv += escapeCSVField(prog.Num) + "," + escapeCSVField(prog.Name) + "," + escapeCSVField(prog.Person) + "\n"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"csv": csv})
}

// CSV Import for Lyrics
func importLyricsCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	var input struct {
		CSV string `json:"csv"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse CSV (expects: text,transition_time per line)
	lines := parseCSVLines(input.CSV)
	if len(lines) == 0 {
		http.Error(w, "Empty CSV", http.StatusBadRequest)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// Delete existing lyrics
		if err := tx.Where("song_id = ?", songID).Delete(&Lyric{}).Error; err != nil {
			return err
		}

		// Create new lyrics
		for i, line := range lines {
			fields := parseCSVLine(line)
			if len(fields) < 1 {
				continue // Skip empty lines
			}

			text := fields[0]
			transitionTime := 1.0
			if len(fields) >= 2 {
				if tt, err := strconv.ParseFloat(fields[1], 64); err == nil {
					transitionTime = tt
				}
			}

			lyric := Lyric{
				SongID:         uint(songID),
				Text:           text,
				TransitionTime: transitionTime,
				Position:       i,
			}

			if err := tx.Create(&lyric).Error; err != nil {
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

// CSV Export for Lyrics
func exportLyricsCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	var lyrics []Lyric
	if err := db.Where("song_id = ?", songID).Order("position").Find(&lyrics).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate CSV
	csv := ""
	for _, lyric := range lyrics {
		csv += escapeCSVField(lyric.Text) + "," + strconv.FormatFloat(lyric.TransitionTime, 'f', 1, 64) + "\n"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"csv": csv})
}

// JSON Export for Key
func exportKeyJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var key Key
	if err := db.First(&key, keyID).Error; err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	keyResp := buildKeyResponse(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keyResp)
}

// JSON Import for Key
func importKeyJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var input KeyResponse

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// Update key metadata
		var key Key
		if err := tx.First(&key, keyID).Error; err != nil {
			return err
		}

		key.Name = input.Name
		key.KeyType = input.KeyType
		key.TransitionTime = input.TransitionTime

		if err := tx.Save(&key).Error; err != nil {
			return err
		}

		if key.KeyType == "program" {
			// Delete existing programs
			if err := tx.Where("key_id = ?", keyID).Delete(&Program{}).Error; err != nil {
				return err
			}

			// Create new programs
			for _, prog := range input.Programs {
				program := Program{
					KeyID:    uint(keyID),
					Num:      prog.Num,
					Name:     prog.Name,
					Person:   prog.Person,
					Position: prog.Position,
				}
				if err := tx.Create(&program).Error; err != nil {
					return err
				}
			}
		} else if key.KeyType == "lyrics" {
			// Delete existing songs
			if err := tx.Where("key_id = ?", keyID).Delete(&Song{}).Error; err != nil {
				return err
			}

			// Create new songs
			for _, song := range input.Songs {
				newSong := Song{
					KeyID:    uint(keyID),
					Name:     song.Name,
					Position: song.Position,
				}
				if err := tx.Create(&newSong).Error; err != nil {
					return err
				}

				// Create lyrics
				for _, lyric := range song.Lyrics {
					newLyric := Lyric{
						SongID:         newSong.ID,
						Text:           lyric.Text,
						TransitionTime: lyric.TransitionTime,
						Position:       lyric.Position,
					}
					if err := tx.Create(&newLyric).Error; err != nil {
						return err
					}
				}
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

// CSV helper functions
func parseCSVLines(csv string) []string {
	lines := []string{}
	currentLine := ""
	inQuotes := false

	for i := 0; i < len(csv); i++ {
		c := csv[i]
		if c == '"' {
			inQuotes = !inQuotes
		} else if (c == '\n' || c == '\r') && !inQuotes {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
		} else {
			currentLine += string(c)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func parseCSVLine(line string) []string {
	fields := []string{}
	currentField := ""
	inQuotes := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				currentField += "\""
				i++
			} else {
				inQuotes = !inQuotes
			}
		} else if c == ',' && !inQuotes {
			fields = append(fields, currentField)
			currentField = ""
		} else {
			currentField += string(c)
		}
	}

	fields = append(fields, currentField)
	return fields
}

func escapeCSVField(field string) string {
	needsQuotes := false
	for _, c := range field {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}

	if needsQuotes {
		escaped := "\""
		for _, c := range field {
			if c == '"' {
				escaped += "\"\""
			} else {
				escaped += string(c)
			}
		}
		escaped += "\""
		return escaped
	}

	return field
}
