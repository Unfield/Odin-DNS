package odintypes

type DNSRequest struct {
	Header    HSection
	Questions []Question
}

type HSection struct {
	ID      uint16
	Flags   Flags
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type Flags struct {
	QR     bool
	Opcode int
	AA     bool
	TC     bool
	RD     bool
	RA     bool
	Z      bool
	RCode  int
}

type Question struct {
	Name  string `json:"name" yaml:"name" xml:"name"`
	Type  string `json:"type" yaml:"type" xml:"type"`
	Class string `json:"class" yaml:"class" xml:"class"`
}

type ResourceRecord struct {
	Name     string `json:"name" yaml:"name" xml:"name"`
	Type     string `json:"type" yaml:"type" xml:"type"`
	Class    string `json:"class" yaml:"class" xml:"class"`
	TTL      int    `json:"ttl" yaml:"ttl" xml:"ttl"`
	RDLength int    `json:"rdlength" yaml:"rdlength" xml:"rdlength"`
	RData    string `json:"rdata" yaml:"rdata" xml:"rdata"`
}

type Zone struct {
	Name     string            `json:"name" yaml:"name" xml:"name"`
	Records  []ResourceRecord  `json:"records" yaml:"records" xml:"records"`
	Metadata map[string]string `json:"metadata" yaml:"metadata" xml:"metadata"`
}
