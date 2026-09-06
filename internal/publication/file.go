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
