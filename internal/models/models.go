package models

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Exercise struct {
	ID              string
	Name            string
	BodyPart        *string // TODO: enum
	PrimaryMuscle   *string // TODO: enum
	SecondaryMuscle *string // TODO: enum
	IsActive        bool
}

type Session struct {
	ID          string
	UserID      string
	PerformedAt time.Time
	Notes       *string
	SessionType *string // TODO: enum
	Sets        []Set
}

type Set struct {
	ID         string
	SessionID  string
	ExerciseID string
	SetIndex   int
	Reps       int
	WeightKG   float64
	RPE        *int
}

// PlanSuggestion is a lightweight container for the progression endpoint.
type PlanSuggestion struct {
	ExerciseID string  `json:"exercise_id"`
	WeightKG   float64 `json:"weight_kg"`
	Reps       int     `json:"reps"`
	Notes      string  `json:"notes,omitempty"`
}
