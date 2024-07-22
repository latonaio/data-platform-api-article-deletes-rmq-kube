package requests

type Header struct {
	Article				int     `json:"Article"`
	IsMarkedForDeletion	*bool	`json:"IsMarkedForDeletion"`
}
