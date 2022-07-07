package config

import (
	"bytes"
	_ "embed"
	"html/template"
	"time"

	"github.com/pkg/errors"
)

//go:embed genesis/genesis.tmpl.json
var genesisTemplate string

func genesis(n Network) ([]byte, error) {
	genesisBuf := new(bytes.Buffer)
	err := template.Must(template.New("genesis").Parse(genesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC string
		ChainID        ChainID
		TokenSymbol    string
	}{
		GenesisTimeUTC: n.genesisTime.UTC().Format(time.RFC3339),
		ChainID:        n.chainID,
		TokenSymbol:    n.tokenSymbol,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}
	return genesisBuf.Bytes(), nil
}
