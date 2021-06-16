package models

// file
type File struct {
	FID      int
	FileName string `json:"fileName"`
	FileType string `json:"fileType"`
	URL      string `json:"url"`
	Content  string `json:"content"`
}
