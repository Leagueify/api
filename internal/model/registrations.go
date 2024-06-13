package model

import "github.com/lib/pq"

type Registration struct {
	ID         string
	PlayerIDs  pq.StringArray
	AmountDue  int
	AmountPaid int
}
