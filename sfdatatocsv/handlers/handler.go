package handlers

import (
	"encoding/json"
	"io"

	"go.uber.org/zap"
)

func GetHandler(l *zap.Logger, cfg *Config) (*Handler, error) {
	return &Handler{
		l:       l,
		cfg:     cfg,
		mcIdMap: map[string]MeasureCalcRecords{},
	}, nil
}

func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func FromJSON(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}
