package postgres

import (
	"errors"

	"github.com/lib/pq"
)

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == "23505"
}
