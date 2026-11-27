package servefile

import (
	"github.com/utmstack/utmstack/backend/modules/servefile/handler"
	"github.com/utmstack/utmstack/backend/modules/servefile/repository"
	"github.com/utmstack/utmstack/backend/modules/servefile/usecase"
	"gorm.io/gorm"
)

type Module struct {
	fileHandler *handler.FileHandler
}

func NewModule(db *gorm.DB, encryptionKey, storageDir string) *Module {
	repo := repository.NewServedFileRepository(db)
	uc := usecase.NewServedFileUsecase(repo, storageDir, encryptionKey)
	return &Module{fileHandler: handler.NewFileHandler(uc)}
}

func (m *Module) GetFileHandler() *handler.FileHandler { return m.fileHandler }
