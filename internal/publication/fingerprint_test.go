package publication

import (
	"crypto/sha256"
	"testing"
)

func TestCalculateContentFingerprintIsDeterministic(t *testing.T) {
	firstDigest := sha256.Sum256([]byte("primeiro arquivo"))
	secondDigest := sha256.Sum256([]byte("segundo arquivo"))

	firstFile := VerifiedFile{
		ClassifiedFile: ClassifiedFile{
			DatasetCode: "companies",
			PartNumber:  0,
		},
		SHA256: firstDigest,
	}

	secondFile := VerifiedFile{
		ClassifiedFile: ClassifiedFile{
			DatasetCode: "companies",
			PartNumber:  1,
		},
		SHA256: secondDigest,
	}

	firstResult, err := CalculateContentFingerprint(
		[]VerifiedFile{firstFile, secondFile},
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	secondResult, err := CalculateContentFingerprint(
		[]VerifiedFile{secondFile, firstFile},
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	if firstResult != secondResult {
		t.Errorf(
			"fingerprint deveria ser independente da ordem dos arquivos",
		)
	}

	combinedDigests := make([]byte, 0, sha256.Size*2)
	combinedDigests = append(combinedDigests, firstDigest[:]...)
	combinedDigests = append(combinedDigests, secondDigest[:]...)

	expected := sha256.Sum256(combinedDigests)

	if firstResult != expected {
		t.Errorf(
			"fingerprint deveria ser %x, mas recebeu %x",
			expected,
			firstResult,
		)
	}
}

func TestCalculateContentFingerprintRejectsEmptyFiles(t *testing.T) {
	result, err := CalculateContentFingerprint(nil)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != ([sha256.Size]byte{}) {
		t.Errorf(
			"fingerprint deveria estar vazio, mas recebeu %x",
			result,
		)
	}
}
