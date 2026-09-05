package publication

type PartNumberSource string

const (
	PartNumberSourceFixed        PartNumberSource = "fixed"
	PartNumberSourceCaptureGroup PartNumberSource = "capture_group"
)

type Catalog struct {
	FormatVersion uint16    `json:"format_version"`
	Datasets      []Dataset `json:"datasets"`
}

type Dataset struct {
	Code           string         `json:"code"`
	FilePattern    string         `json:"file_pattern"`
	PartNumberRule PartNumberRule `json:"part_number"`
}

type PartNumberRule struct {
	Source       PartNumberSource `json:"source"`
	Value        uint16           `json:"value,omitempty"`
	CaptureGroup uint8            `json:"capture_group,omitempty"`
}
