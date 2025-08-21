package pathmatcher

import (
	"fmt"
	"strings"
)

type path struct {
	Template string
	Segments []segment
}

type segment struct {
	IsParam bool
	Param   string
	Const   string
}

type MatchingError struct {
	path        *path
	description string
}

func newMatchingError(path *path, description string) MatchingError {
	return MatchingError{path, description}
}

func (e MatchingError) Error() string {
	return fmt.Sprintf("matching template %q: %s", e.path.Template, e.description)
}

func NewPath(template string) *path {
	stringSegments := strings.Split(template, "/")
	segments := make([]segment, len(stringSegments))

	for index, stringSegment := range stringSegments {
		if strings.HasPrefix(stringSegment, "{") && strings.HasSuffix(stringSegment, "}") {
			segments[index] = segment{IsParam: true, Param: stringSegment[1 : len(stringSegment)-1]}
		} else {
			segments[index] = segment{IsParam: false, Const: stringSegment}
		}
	}

	return &path{
		Template: template,
		Segments: segments,
	}
}

func (p *path) Match(path string) (map[string]string, error) {
	stringSegments := strings.Split(path, "/")

	if len(stringSegments) != len(p.Segments) {
		return nil, newMatchingError(p, fmt.Sprintf("path %q does not match", path))
	}

	params := make(map[string]string)

	for index, stringSegment := range stringSegments {
		templateSegment := p.Segments[index]

		if templateSegment.IsParam {
			if stringSegment == "" {
				return nil, newMatchingError(p, fmt.Sprintf("path %q has an empty segment", path))
			}

			params[templateSegment.Param] = stringSegment
		} else if p.Segments[index].Const != stringSegment {
			return nil, newMatchingError(p, fmt.Sprintf("path %q does not match", path))
		}
	}

	return params, nil
}
