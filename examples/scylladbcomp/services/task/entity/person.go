package entity

import "github.com/scylladb/gocqlx/table"

// metadata specifies table name and columns it must be in sync with schema.
var personMetadata = table.Metadata{
	Name:    "person",
	Columns: []string{"first_name", "last_name", "email"},
	PartKey: []string{"first_name"},
	SortKey: []string{"last_name"},
}

// PersonTable allows for simple CRUD operations based on personMetadata.
var PersonTable = table.New(personMetadata)

// Person represents a row in person table.
// Field names are converted to snake case by default, no need to add special tags.
// A field will not be persisted by adding the `db:"-"` tag or making it unexported.
type Person struct {
	FirstName string
	LastName  string
	Email     []string
	HairColor string `db:"-"` // exported and skipped
	eyeColor  string // unexported also skipped
}

// PersonCreateRequest represents the request payload for creating a person
type PersonCreateRequest struct {
	FirstName string   `json:"first_name" binding:"required"`
	LastName  string   `json:"last_name" binding:"required"`
	Email     []string `json:"email"`
}

// PersonFilter represents the filter for listing persons
type PersonFilter struct {
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
}
