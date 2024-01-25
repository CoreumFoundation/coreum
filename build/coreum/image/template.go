package image

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed Dockerfile.tmpl
	tmpl       string
	dockerfile = template.Must(template.New("dockerfile").Parse(tmpl))
)

// Data is the structure containing fields required by the template.
type Data struct {
	// From is the tag of the base image
	From string

	// CoredBinary is the name of cored binary file to copy from build context
	CoredBinary string

	// CosmovisorBinary is the name of cosmovisor binary file to copy from build context
	CosmovisorBinary string

	// Networks is the list of available networks
	Networks []string
}

// Execute executes dockerfile template and returns complete dockerfile.
func Execute(data Data) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := dockerfile.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
