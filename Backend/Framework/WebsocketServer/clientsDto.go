package main

type AddBinaryRequest struct {
	FileType string `json:"fileType"`
	RunCmd   string `json:"runCmd"`
	Name     string `json:"name"`
	Content  []byte `json:"content"`
}
