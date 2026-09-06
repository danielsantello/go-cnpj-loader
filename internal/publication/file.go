package publication

type DiscoveredFile struct {
	SourceName     string
	SourceLocation string
}

type ClassifiedFile struct {
	DatasetCode    string
	PartNumber     uint16
	SourceName     string
	SourceLocation string
}

type VerifiedFile struct {
	ClassifiedFile ClassifiedFile
	SizeBytes      uint64
	SHA256         [32]byte
}
