package models

type EditGalleryPageData struct {
	Title             string
	ActiveLink        string
	Gallery           *Gallery
	Media             []*Media
	Page              int
	Limit             int
	MediaCount        int
	TotalPages        int
	HasNext           bool
	PaginationBaseURL string
	Target            string
}
