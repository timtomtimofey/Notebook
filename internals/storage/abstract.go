package storage

import "context"

// if some field has "" val then it is considered unset
type Note struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Company   *string `json:"company,omitempty"`
	Phone     string  `json:"phone"`
	Mail      string  `json:"mail"`
	BirthDate *string `json:"birthdate,omitempty"`
	// TODO: Photo
}

func (n *Note) VerifyNote() bool {
	// check that required fields are present:
	// - Name
	// - Phone
	// - Mail
	if n.Name == "" || n.Phone == "" || n.Mail == "" {
		return false
	}
	return true
}

type NoteStorage interface {
	IsNote(ctx context.Context, id string) (bool, error)
	GetNote(ctx context.Context, id string) (Note, error)
	GetRangeNotes(ctx context.Context, offset, limit int) ([]Note, error) // if limit == -1 then all notes are served
	AddNote(ctx context.Context, n *Note) (Note, error)
	UpdateNote(ctx context.Context, id string, n *Note) (Note, error)
	DeleteNote(ctx context.Context, id string) error
}
