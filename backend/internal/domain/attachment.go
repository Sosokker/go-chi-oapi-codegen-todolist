package domain

// Attachment info returned by AddAttachment
type AttachmentInfo struct {
	FileID      string `json:"fileId"`
	FileName    string `json:"fileName"`
	FileURL     string `json:"fileUrl"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}
