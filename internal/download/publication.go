package download

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileProgress func(
	current int,
	total int,
	file File,
)

type PublicationResult struct {
	FileCount       int
	DownloadedCount int
	ReusedCount     int
	TotalSizeBytes  uint64
}

func DownloadPublication(
	ctx context.Context,
	client *http.Client,
	referenceYear uint16,
	referenceMonth uint8,
	destinationDirectory string,
	progress FileProgress,
) (PublicationResult, error) {
	if client == nil {
		return PublicationResult{}, errors.New(
			"cliente HTTP não foi informado",
		)
	}

	if strings.TrimSpace(destinationDirectory) == "" {
		return PublicationResult{}, errors.New(
			"diretório de destino é obrigatório",
		)
	}

	if !filepath.IsAbs(destinationDirectory) {
		return PublicationResult{}, errors.New(
			"diretório de destino deve possuir caminho absoluto",
		)
	}

	if err := os.MkdirAll(
		destinationDirectory,
		0o755,
	); err != nil {
		return PublicationResult{}, fmt.Errorf(
			"não foi possível criar o diretório de destino %q: %w",
			destinationDirectory,
			err,
		)
	}

	directoryInfo, err := os.Stat(destinationDirectory)
	if err != nil {
		return PublicationResult{}, fmt.Errorf(
			"não foi possível consultar o diretório de destino %q: %w",
			destinationDirectory,
			err,
		)
	}

	if !directoryInfo.IsDir() {
		return PublicationResult{}, fmt.Errorf(
			"destino %q não é um diretório",
			destinationDirectory,
		)
	}

	files, err := ListPublicationFiles(
		ctx,
		client,
		referenceYear,
		referenceMonth,
	)
	if err != nil {
		return PublicationResult{}, err
	}

	result := PublicationResult{
		FileCount: len(files),
	}

	for index, file := range files {
		if progress != nil {
			progress(
				index+1,
				len(files),
				file,
			)
		}

		downloaded, err := DownloadFile(
			ctx,
			client,
			referenceYear,
			referenceMonth,
			file,
			destinationDirectory,
		)
		if err != nil {
			return PublicationResult{}, fmt.Errorf(
				"não foi possível obter a publicação: %w",
				err,
			)
		}

		result.TotalSizeBytes += file.SizeBytes

		if downloaded {
			result.DownloadedCount++
		} else {
			result.ReusedCount++
		}
	}

	return result, nil
}
