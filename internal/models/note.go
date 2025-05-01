package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Note struct {
	gorm.Model
	UserID      uint          `gorm:"not null"`
	Title       string        `gorm:"not null"`
	Content     string        `gorm:"not null"`
	Summary     string        // Суммаризация текста (можно генерировать на стороне AI)
	Embedding   []byte        `gorm:"type:bytea"` // Векторное представление заметки
	Attachments []Attachment  // Вложения к заметке
	IsArchived  bool          // Архивная заметка или нет
	Tags        []Tag         `gorm:"many2many:note_tags;"`        // Связь многие-ко-многим с тегами
	RelatedIDs  pq.Int64Array `gorm:"type:integer[];default:'{}'"` // Связанные заметки (ID других заметок)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Topic struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string
}

type Tag struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string
	Notes       []Note `gorm:"many2many:note_tags;"` // Обратная связь с заметками
}

type Attachment struct {
	gorm.Model
	NoteID     uint   `gorm:"not null"`
	FileURL    string `gorm:"not null"` // Путь до файла
	FileType   string // Тип файла (например, "image", "audio", "pdf")
	FileSize   int64  // Размер файла в байтах
	UploadedAt time.Time
}
