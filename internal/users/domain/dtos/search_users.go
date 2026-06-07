package dtos

// SearchUsersRequest represents the search criteria for users
type SearchUsersRequest struct {
	Q      string `query:"q" json:"q"`
	Limit  int    `query:"limit" json:"limit"`
	Offset int    `query:"offset" json:"offset"`
	Activo *bool  `query:"activo" json:"activo"`
}
