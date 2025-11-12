package main

import (
	"log"
	"time"
)

// handleStateTransition manages server-side state transitions
func handleStateTransition(keyID uint, newStatus string, transitionTime float64, keyType string) {
	switch newStatus {
	case "OPENING":
		// OPENING -> OPENED after transition_time
		duration := time.Duration(transitionTime * float64(time.Second))
		stateManager.Start(keyID, duration, "OPENING", "OPENED", func() {
			updateKeyStatus(keyID, "OPENED")
		})

	case "CLOSING":
		// CLOSING -> CLOSED after transition_time
		duration := time.Duration(transitionTime * float64(time.Second))
		stateManager.Start(keyID, duration, "CLOSING", "CLOSED", func() {
			updateKeyStatus(keyID, "CLOSED")
		})

	case "PLAYING_FORWARD":
		// Play forward one lyric, then return to OPENED
		if keyType == "lyrics" {
			go playForwardOneLyric(keyID)
		}

	case "OPENED", "CLOSED":
		// Cancel any ongoing transitions
		stateManager.Cancel(keyID)
	}
}

// updateKeyStatus updates a key's status and broadcasts the change
func updateKeyStatus(keyID uint, newStatus string) {
	var key Key
	if err := db.First(&key, keyID).Error; err != nil {
		log.Printf("Error finding key %d: %v", keyID, err)
		return
	}

	key.Status = newStatus
	key.Version = key.Version + 1

	if err := db.Save(&key).Error; err != nil {
		log.Printf("Error updating key %d status: %v", keyID, err)
		return
	}

	log.Printf("Key %d status updated: %s", keyID, newStatus)
	broadcastUpdate()
}

// playForwardOneLyric plays forward one lyric and returns to OPENED status
func playForwardOneLyric(keyID uint) {
	var key Key
	if err := db.First(&key, keyID).Error; err != nil {
		log.Printf("Error finding key %d: %v", keyID, err)
		return
	}

	// Get current song
	if key.CurrentPresetID == nil {
		updateKeyStatus(keyID, "OPENED")
		return
	}

	var song Song
	if err := db.First(&song, *key.CurrentPresetID).Error; err != nil {
		log.Printf("Error finding song %d: %v", *key.CurrentPresetID, err)
		updateKeyStatus(keyID, "OPENED")
		return
	}

	// Get current lyric
	if song.CurrentLyricID == nil {
		updateKeyStatus(keyID, "OPENED")
		return
	}

	var currentLyric Lyric
	if err := db.First(&currentLyric, *song.CurrentLyricID).Error; err != nil {
		log.Printf("Error finding lyric %d: %v", *song.CurrentLyricID, err)
		updateKeyStatus(keyID, "OPENED")
		return
	}

	// Wait for the current lyric's transition time
	duration := time.Duration(currentLyric.TransitionTime * float64(time.Second))
	time.Sleep(duration)

	// Check status again after sleep (in case it was changed)
	if err := db.First(&key, keyID).Error; err != nil || key.Status != "PLAYING_FORWARD" {
		return
	}

	// Return to OPENED status after playing forward one lyric
	updateKeyStatus(keyID, "OPENED")
	log.Printf("Played forward one lyric for key %d, returning to OPENED", keyID)
}
