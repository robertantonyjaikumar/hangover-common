package utils

import (
	"github.com/robertantonyjaikumar/hangover-common/config"
	"github.com/robertantonyjaikumar/hangover-common/logger"
	"go.uber.org/zap"
	"os"
)

func ReadJSON(fileName string) (File *os.File, err error) {
	filePath := config.CFG.V.Get("seed_path").(string) + fileName
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("Error opening JSON file", zap.Error(err))
		return nil, err
	}
	defer file.Close()

	return file, nil
}
