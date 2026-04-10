package client

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Owner string `json:"owner"`
}

type DeleteTagRequest struct {
	ID string `json:"id"`
}

type RenameTagRequest struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type UpdateTagColorRequest struct {
	Tag   string `json:"tag"`
	Color string `json:"color"`
}
