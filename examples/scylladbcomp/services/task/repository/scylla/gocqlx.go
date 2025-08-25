package scylla

import (
	"context"

	"github.com/scylladb/gocqlx/qb"
	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"
)

func (repo *scyllaRepo) AddNewPerson(ctx context.Context, person *entity.Person) error {
	q := repo.sessionWithGocqlX.Query(entity.PersonTable.Insert()).BindStruct(person)
	if err := q.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (repo *scyllaRepo) ListPersons(ctx context.Context, firstName, lastName string) (*[]entity.Person, error) {
	var people []entity.Person

	// Build query based on provided filters
	if firstName != "" && lastName != "" {
		// If both firstName and lastName are provided, use both
		q := repo.sessionWithGocqlX.Query(entity.PersonTable.Select()).BindMap(qb.M{
			"first_name": firstName,
			"last_name":  lastName,
		})
		if err := q.SelectRelease(&people); err != nil {
			return nil, err
		}
	} else if firstName != "" {
		// If only firstName is provided (partition key)
		q := repo.sessionWithGocqlX.Query(entity.PersonTable.Select()).BindMap(qb.M{
			"first_name": firstName,
		})
		if err := q.SelectRelease(&people); err != nil {
			return nil, err
		}
	} else {
		// If no filters, return all persons using raw CQL query
		q := repo.sessionWithGocqlX.Query("SELECT first_name, last_name, email FROM person", []string{})
		if err := q.SelectRelease(&people); err != nil {
			return nil, err
		}
	}

	return &people, nil
}
