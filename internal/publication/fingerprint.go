package publication

import (
	"crypto/sha256"
	"errors"
	"sort"
)

func CalculateContentFingerprint(
	files []VerifiedFile,
) ([sha256.Size]byte, error) {
	if len(files) == 0 {
		return [sha256.Size]byte{}, errors.New(
			"não existem arquivos verificados para calcular o fingerprint",
		)
	}

	orderedFiles := append([]VerifiedFile(nil), files...)

	sort.Slice(
		orderedFiles,
		func(left int, right int) bool {
			leftFile := orderedFiles[left].ClassifiedFile
			rightFile := orderedFiles[right].ClassifiedFile

			if leftFile.DatasetCode != rightFile.DatasetCode {
				return leftFile.DatasetCode < rightFile.DatasetCode
			}

			return leftFile.PartNumber < rightFile.PartNumber
		},
	)

	hasher := sha256.New()

	for _, file := range orderedFiles {
		hasher.Write(file.SHA256[:])
	}

	var result [sha256.Size]byte
	copy(result[:], hasher.Sum(nil))

	return result, nil
}
