package Country

import "github.com/google/uuid"

type Country struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"ShortName"`
}

func Countries() map[string]string {

	return nil
}
