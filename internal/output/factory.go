package output

import (
	"fmt"
)

// GetFormatter returns the appropriate formatter based on format type
func GetFormatter(format Format) (Formatter, error) {
	switch format {
	case FormatPlain:
		return NewPlainFormatter(), nil
	case FormatJSON:
		return NewJSONFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
