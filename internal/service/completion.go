package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ops-platform/internal/model"
	"ops-platform/internal/repository"
)

const (
	maxFileSize   = 50 * 1024 * 1024 // 50MB
	uploadDir     = "uploads"
)

var allowedExtensions = map[string]bool{
	".txt": true, ".pdf": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	".csv": true, ".json": true, ".xml": true, ".yaml": true, ".yml": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
	".log": true, ".conf": true, ".cfg": true, ".ini": true,
	".sh": true, ".py": true, ".go": true, ".sql": true, ".md": true,
}

type CompletionService struct {
	completionRepo *repository.CompletionRepository
}

func NewCompletionService(completionRepo *repository.CompletionRepository) *CompletionService {
	return &CompletionService{completionRepo: completionRepo}
}

func (s *CompletionService) SaveCompletion(c *model.TicketCompletion) error {
	return s.completionRepo.Create(c)
}

func (s *CompletionService) GetCompletion(ticketID int64) (*model.TicketCompletion, error) {
	return s.completionRepo.FindByTicketID(ticketID)
}

func (s *CompletionService) SaveFile(ticketID, uploaderID int64, filename string, reader io.Reader) (*model.TicketFile, error) {
	// Sanitize: prevent path traversal
	cleanName := filepath.Base(filename)
	if cleanName == "." || cleanName == "/" {
		return nil, errors.New("invalid filename")
	}

	// Whitelist check
	ext := strings.ToLower(filepath.Ext(cleanName))
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("file type %s not allowed", ext)
	}

	os.MkdirAll(uploadDir, 0755)

	// Unique stored name: ticketID_timestamp.ext (no overwrite possible)
	storedName := fmt.Sprintf("%d_%d%s", ticketID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, storedName)

	dst, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Limit copy to maxFileSize
	limitedReader := io.LimitReader(reader, maxFileSize+1)
	size, err := io.Copy(dst, limitedReader)
	if err != nil {
		os.Remove(savePath)
		return nil, err
	}
	if size > maxFileSize {
		os.Remove(savePath)
		return nil, fmt.Errorf("file size exceeds limit (max %dMB)", maxFileSize/(1024*1024))
	}

	f := &model.TicketFile{
		TicketID:   ticketID,
		Filename:   cleanName,
		Filepath:   savePath,
		Filesize:   size,
		UploaderID: uploaderID,
	}
	if err := s.completionRepo.CreateFile(f); err != nil {
		os.Remove(savePath)
		return nil, err
	}
	return f, nil
}

func (s *CompletionService) GetFiles(ticketID int64) ([]model.TicketFile, error) {
	return s.completionRepo.FindFilesByTicketID(ticketID)
}

func (s *CompletionService) GetFile(fileID int64) (*model.TicketFile, error) {
	return s.completionRepo.FindFileByID(fileID)
}

func (s *CompletionService) DeleteFile(fileID int64) error {
	f, err := s.completionRepo.FindFileByID(fileID)
	if err != nil {
		return err
	}
	os.Remove(f.Filepath)
	return s.completionRepo.DeleteFile(fileID)
}

// DeleteAllFiles removes all files for a ticket (used when deleting ticket)
func (s *CompletionService) DeleteAllFiles(ticketID int64) error {
	files, err := s.completionRepo.FindFilesByTicketID(ticketID)
	if err != nil {
		return nil
	}
	for _, f := range files {
		os.Remove(f.Filepath)
		s.completionRepo.DeleteFile(f.ID)
	}
	return nil
}
