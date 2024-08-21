package fileup

// 参考：
// https://github.com/tsawler/toolbox

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const defaultMaxUpload = 10485760 // 10mb

type UploadLocalFile struct {
	MaxFileSize        int      // 上传文件最大限制
	AllowedFileTypes   []string // 允许上次类型 (e.g. image/jpeg)
	AllowUnknownFields bool     // 是否允许上传未知类型文件
	SaveDir            string
}

// NewUploadLocalFile 创建一个本地上传实例
func NewUploadLocalFile(
	maxFileSize int,
	allowedFileTypes []string,
	allowUnknownFields bool,
	saveDir string,
) *UploadLocalFile {
	return &UploadLocalFile{
		MaxFileSize:        maxFileSize,
		AllowedFileTypes:   allowedFileTypes,
		AllowUnknownFields: allowUnknownFields,
		SaveDir:            saveDir,
	}
}

// UploadedFile 上传后返回的结果结构
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// UploadOneFile 单个文件上传，最终时是调用 UploadFiles 上传
func (t *UploadLocalFile) UploadOneFile(r *http.Request, rename bool) (*UploadedFile, error) {
	files, err := t.UploadFiles(r, rename)
	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// UploadFiles 执行上传，保存文件的实际方法，支持多文件上传
func (t *UploadLocalFile) UploadFiles(r *http.Request, rename bool) ([]*UploadedFile, error) {

	var uploadedFiles []*UploadedFile

	// 如果上传目录不存在，请创建该目录
	err := t.CreateDirIfNotExist(t.SaveDir)
	if err != nil {
		return nil, err
	}

	if t.MaxFileSize <= 0 {
		t.MaxFileSize = defaultMaxUpload
	}

	// 解析form表单，同时限制文件大小
	err = r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, fmt.Errorf("error parsing form data: %v", err)
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				if hdr.Size > int64(t.MaxFileSize) {
					return nil, fmt.Errorf("the uploaded file is too big, and must be less than %d", t.MaxFileSize)
				}

				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				allowed := false
				filetype := http.DetectContentType(buff)
				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(filetype, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("the uploaded file type is not permitted")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}

				randomStr, err := uuid.NewRandom()
				if err != nil {
					return nil, err
				}

				if rename {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", randomStr, filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}
				uploadedFile.OriginalFileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(t.SaveDir, uploadedFile.NewFileName)); nil != err {
					return nil, err
				}
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}

// CreateDirIfNotExist 如果目录不存在，则创建目录和所有必要的父目录。
func (t *UploadLocalFile) CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
