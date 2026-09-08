package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFile(
	ctx context.Context,
	client *http.Client,
	referenceYear uint16,
	referenceMonth uint8,
	file File,
	destinationDirectory string,
) (bool, error) {
	if client == nil {
		return false, errors.New("cliente HTTP não foi informado")
	}

	if referenceYear < 1000 || referenceYear > 9999 {
		return false, fmt.Errorf(
			"ano de referência é inválido: %d",
			referenceYear,
		)
	}

	if referenceMonth < 1 || referenceMonth > 12 {
		return false, fmt.Errorf(
			"mês de referência é inválido: %d",
			referenceMonth,
		)
	}

	if strings.TrimSpace(file.Name) == "" {
		return false, errors.New("nome do arquivo remoto é obrigatório")
	}

	if filepath.Base(file.Name) != file.Name {
		return false, fmt.Errorf(
			"nome do arquivo remoto é inválido: %q",
			file.Name,
		)
	}

	if !strings.EqualFold(filepath.Ext(file.Name), ".zip") {
		return false, fmt.Errorf(
			"arquivo remoto %q não possui extensão ZIP",
			file.Name,
		)
	}

	if file.SizeBytes == 0 {
		return false, fmt.Errorf(
			"arquivo remoto %q possui tamanho inválido",
			file.Name,
		)
	}

	if strings.TrimSpace(destinationDirectory) == "" {
		return false, errors.New(
			"diretório de destino é obrigatório",
		)
	}

	if !filepath.IsAbs(destinationDirectory) {
		return false, errors.New(
			"diretório de destino deve possuir caminho absoluto",
		)
	}

	destinationPath := filepath.Join(
		destinationDirectory,
		file.Name,
	)

	fileInfo, err := os.Stat(destinationPath)
	switch {
	case err == nil:
		if !fileInfo.Mode().IsRegular() {
			return false, fmt.Errorf(
				"destino %q não é um arquivo regular",
				destinationPath,
			)
		}

		if uint64(fileInfo.Size()) != file.SizeBytes {
			return false, fmt.Errorf(
				"arquivo existente %q possui %d bytes; esperava %d",
				destinationPath,
				fileInfo.Size(),
				file.SizeBytes,
			)
		}

		return false, nil

	case !errors.Is(err, os.ErrNotExist):
		return false, fmt.Errorf(
			"não foi possível consultar o destino %q: %w",
			destinationPath,
			err,
		)
	}

	endpoint := fmt.Sprintf(
		"%s/%04d-%02d/%s",
		webDAVBaseURL,
		referenceYear,
		referenceMonth,
		url.PathEscape(file.Name),
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf(
			"não foi possível criar o download de %q: %w",
			file.Name,
			err,
		)
	}

	request.SetBasicAuth(publicShareKey, "")

	response, err := client.Do(request)
	if err != nil {
		return false, fmt.Errorf(
			"não foi possível baixar %q: %w",
			file.Name,
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf(
			"download de %q retornou HTTP %d",
			file.Name,
			response.StatusCode,
		)
	}

	partialPath := destinationPath + ".part"

	partialFile, err := os.OpenFile(
		partialPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		return false, fmt.Errorf(
			"não foi possível criar o arquivo temporário %q: %w",
			partialPath,
			err,
		)
	}

	bytesWritten, copyErr := io.Copy(
		partialFile,
		response.Body,
	)
	closeErr := partialFile.Close()

	if copyErr != nil {
		os.Remove(partialPath)

		return false, fmt.Errorf(
			"não foi possível gravar o arquivo %q: %w",
			file.Name,
			copyErr,
		)
	}

	if closeErr != nil {
		os.Remove(partialPath)

		return false, fmt.Errorf(
			"não foi possível fechar o arquivo temporário %q: %w",
			partialPath,
			closeErr,
		)
	}

	if uint64(bytesWritten) != file.SizeBytes {
		os.Remove(partialPath)

		return false, fmt.Errorf(
			"download de %q possui %d bytes; esperava %d",
			file.Name,
			bytesWritten,
			file.SizeBytes,
		)
	}

	if err := os.Rename(
		partialPath,
		destinationPath,
	); err != nil {
		os.Remove(partialPath)

		return false, fmt.Errorf(
			"não foi possível concluir o arquivo %q: %w",
			file.Name,
			err,
		)
	}

	return true, nil
}
