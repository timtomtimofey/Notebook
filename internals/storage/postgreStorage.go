package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

type PostgreStorage struct {
	conn *pgx.Conn
}

func DBInit(ctx context.Context, connStr string) error {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	if f, err := os.Open("./internals/storage/init.sql"); err != nil {
		return err
	} else if bytes, err := io.ReadAll(f); err != nil {
		return err
	} else {
		_, err = conn.Exec(ctx, string(bytes))
		return err
	}
}

func New(ctx context.Context, connStr string) (*PostgreStorage, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &PostgreStorage{conn}, nil
}

func (ps *PostgreStorage) Close() error {
	return ps.conn.Close(context.Background())
}

func (ps *PostgreStorage) IsNote(ctx context.Context, id string) (bool, error) {
	var exist bool
	err := ps.conn.QueryRow(ctx, "SELECT exists(SELECT 1 FROM notes WHERE id = $1);", id).Scan(&exist)
	return exist, err
}

func rowToNote(row pgx.CollectableRow) (Note, error) {
	var res Note
	err := row.Scan(&res.ID, &res.Name, &res.Company, &res.Phone, &res.Mail, &res.BirthDate, &res.ImageID)
	return res, err
}

func (ps *PostgreStorage) GetNote(ctx context.Context, id string) (Note, error) {
	rows, err := ps.conn.Query(ctx, "SELECT * FROM notes WHERE id = $1;", id)
	if err != nil {
		return Note{}, err
	}
	defer rows.Close()
	return pgx.CollectOneRow(rows, rowToNote)
}

func (ps *PostgreStorage) GetRangeNotes(ctx context.Context, offset, limit int) ([]Note, error) {
	strOffset := strconv.Itoa(offset)
	var strLimit string
	if limit < 0 {
		strLimit = "ALL"
	} else {
		strLimit = strconv.Itoa(limit)
	}
	rows, err := ps.conn.Query(ctx, "SELECT * FROM notes ORDER BY id LIMIT "+strLimit+" OFFSET "+strOffset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, rowToNote)
}

func (ps *PostgreStorage) AddNote(ctx context.Context, n *Note) (Note, error) {
	sql := "INSERT INTO notes (id, full_name, company, phone, mail, birth_date, image_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;"
	rows, err := ps.conn.Query(ctx, sql, n.ID, n.Name, n.Company, n.Phone, n.Mail, n.BirthDate, n.ImageID)
	if err != nil {
		return Note{}, err
	}
	defer rows.Close()
	// log.Print("here", n)
	return pgx.CollectOneRow(rows, rowToNote)
}

func (ps *PostgreStorage) UpdateNote(ctx context.Context, id string, n *Note) (Note, error) {
	updates := make([]string, 0, 6)
	if n.ID != "" {
		updates = append(updates, fmt.Sprintf("id = '%s'", n.ID))
	}
	if n.Name != "" {
		updates = append(updates, fmt.Sprintf("full_name = '%s'", n.Name))
	}
	if n.Company != nil {
		updates = append(updates, fmt.Sprintf("company = '%s'", *n.Company))
	}
	if n.Phone != "" {
		updates = append(updates, fmt.Sprintf("phone = '%s'", n.Phone))
	}
	if n.Mail != "" {
		updates = append(updates, fmt.Sprintf("mail = '%s'", n.Mail))
	}
	if n.BirthDate != nil {
		updates = append(updates, fmt.Sprintf("birth_date = '%s'", *n.BirthDate))
	}
	if n.ImageID != nil {
		updates = append(updates, fmt.Sprintf(" image_id = '%s'", *n.ImageID))
	}
	if len(updates) == 0 {
		return ps.GetNote(ctx, id)
	}

	sql := "UPDATE notes SET " + strings.Join(updates, ", ") + " WHERE id = $1;"
	if _, err := ps.conn.Exec(ctx, sql, id); err != nil {
		return Note{}, err
	}
	if n.ID == "" {
		return ps.GetNote(ctx, id)
	} else {
		return ps.GetNote(ctx, n.ID)
	}
}

func (ps *PostgreStorage) DeleteNote(ctx context.Context, id string) error {
	_, err := ps.conn.Exec(ctx, "DELETE FROM notes WHERE id = $1;", id)
	return err
}
