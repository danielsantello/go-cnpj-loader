package publication

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
)

type zipContentReader struct {
	content io.ReadCloser
	archive *zip.ReadCloser
}

func (reader *zipContentReader) Read(
	buffer []byte,
) (int, error) {
	return reader.content.Read(buffer)
}

func (reader *zipContentReader) Close() error {
	return errors.Join(
		reader.content.Close(),
		reader.archive.Close(),
	)
}

func OpenZIPContent(
	file VerifiedFile,
) (io.ReadCloser, error) {
	sourceName := file.ClassifiedFile.SourceName
	sourceLocation := file.ClassifiedFile.SourceLocation

	archive, err := zip.OpenReader(sourceLocation)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível abrir o ZIP %q: %w",
			sourceName,
			err,
		)
	}

	if len(archive.File) != 1 {
		archive.Close()

		return nil, fmt.Errorf(
			"ZIP %q deveria possuir exatamente um arquivo, mas possui %d",
			sourceName,
			len(archive.File),
		)
	}

	contentFile := archive.File[0]

	if contentFile.FileInfo().IsDir() {
		archive.Close()

		return nil, fmt.Errorf(
			"conteúdo do ZIP %q não é um arquivo regular",
			sourceName,
		)
	}

	if contentFile.UncompressedSize64 == 0 {
		archive.Close()

		return nil, fmt.Errorf(
			"conteúdo do ZIP %q está vazio",
			sourceName,
		)
	}

	content, err := contentFile.Open()
	if err != nil {
		archive.Close()

		return nil, fmt.Errorf(
			"não foi possível abrir o conteúdo do ZIP %q: %w",
			sourceName,
			err,
		)
	}

	return &zipContentReader{
		content: content,
		archive: archive,
	}, nil
}
