package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           *uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Exercise struct {
	ID              *uuid.UUID
	Name            string
	BodyPart        *string // TODO: enum
	PrimaryMuscle   *string // TODO: enum
	SecondaryMuscle *string // TODO: enum
	IsActive        bool
}

type Session struct {
	ID          *uuid.UUID
	UserID      *uuid.UUID
	PerformedAt time.Time
	Notes       *string
	SessionType *string // TODO: enum
	Sets        []Set
}

type Set struct {
	ID         *uuid.UUID
	SessionID  *uuid.UUID
	ExerciseID *uuid.UUID
	SetIndex   int
	Reps       int
	WeightKG   float64
	RPE        *int
}

// PlanSuggestion is a lightweight container for the progression endpoint.
type PlanSuggestion struct {
	ExerciseID *uuid.UUID `json:"exercise_id"`
	WeightKG   float64    `json:"weight_kg"`
	Reps       int        `json:"reps"`
	Notes      string     `json:"notes,omitempty"`
}
