package fileup

import "net/http"

type FileUploader interface {
	UploadOneFile(r *http.Request, rename bool) (*UploadedFile, error)
	UploadFiles(r *http.Request, rename bool) ([]*UploadedFile, error)
}

var _ FileUploader = (*UploadLocalFile)(nil)
