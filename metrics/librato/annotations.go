package librato

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const annotationsURL = "https://metrics-api.librato.com/v1/annotations"

var ErrNoNameAnnotation = fmt.Errorf("Annotation must have name")

// Annotation is a representation of librato annotation object
// https://www.librato.com/docs/kb/visualize/annotations/
type Annotation struct {
	Title       string  `json:"title"`
	Source      *string `json:"source"`
	Description *string `json:"description"`
	Links       []Link  `json:"links"`

	StartTime *int64 `json:"start_time"`
	EndTime   *int64 `json:"end_time"`
}

// Link is a representation of link object, that can be used in annotations
// https://www.librato.com/docs/api/#update-an-annotation
type Link struct {
	Relationship string  `json:"rel"`
	URL          string  `json:"href"`
	Label        *string `json:"label"`
}

// PostAnnotation sends annotation to librato API right away
// because Annotation to doesn't seem to support batching
// http://api-docs-archive.librato.com/#create-an-annotation
func (lb *Librato) PostAnnotation(body *Annotation, name string) error {
	if name == "" {
		return ErrNoNameAnnotation
	}

	b, err := json.Marshal(body)
	if nil != err {
		return err
	}

	return lb.makeRequest(bytes.NewBuffer(b), fmt.Sprintf("%s/%s", annotationsURL, name))
}
