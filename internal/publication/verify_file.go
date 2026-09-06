package publication

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func VerifyFile(
	file ClassifiedFile,
) (VerifiedFile, error) {
	if strings.TrimSpace(file.SourceName) == "" {
		return VerifiedFile{}, errors.New(
			"nome do arquivo classificado é obrigatório",
		)
	}

	if strings.TrimSpace(file.SourceLocation) == "" {
		return VerifiedFile{}, errors.New(
			"localização do arquivo classificado é obrigatória",
		)
	}

	if !filepath.IsAbs(file.SourceLocation) {
		return VerifiedFile{}, errors.New(
			"localização do arquivo classificado deve ser absoluta",
		)
	}

	sourceFile, err := os.Open(file.SourceLocation)
	if err != nil {
		return VerifiedFile{}, fmt.Errorf(
			"não foi possível abrir o arquivo %q: %w",
			file.SourceName,
			err,
		)
	}
	defer sourceFile.Close()

	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return VerifiedFile{}, fmt.Errorf(
			"não foi possível consultar o arquivo %q: %w",
			file.SourceName,
			err,
		)
	}

	if !fileInfo.Mode().IsRegular() {
		return VerifiedFile{}, fmt.Errorf(
			"origem %q não é um arquivo regular",
			file.SourceName,
		)
	}

	if fileInfo.Size() == 0 {
		return VerifiedFile{}, fmt.Errorf(
			"arquivo %q está vazio",
			file.SourceName,
		)
	}

	hasher := sha256.New()

	bytesRead, err := io.Copy(hasher, sourceFile)
	if err != nil {
		return VerifiedFile{}, fmt.Errorf(
			"não foi possível calcular o SHA-256 do arquivo %q: %w",
			file.SourceName,
			err,
		)
	}

	if bytesRead != fileInfo.Size() {
		return VerifiedFile{}, fmt.Errorf(
			"arquivo %q foi alterado durante a verificação",
			file.SourceName,
		)
	}

	var digest [sha256.Size]byte
	copy(digest[:], hasher.Sum(nil))

	return VerifiedFile{
		ClassifiedFile: file,
		SizeBytes:      uint64(bytesRead),
		SHA256:         digest,
	}, nil
}
