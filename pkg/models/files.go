package models

// file
type File struct {
	FID      int
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	URL      string `json:"url"`
	Content  string `json:"content"`
}
