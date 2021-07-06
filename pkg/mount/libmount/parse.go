/*
Copyright 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package libmount

import (
	"errors"
	"strings"
)

var (
	ErrInvalidToken error = errors.New("invalid token error")
)

type Parser interface {
	Parse(string) (*Filesystem, error)
}

type mountInfoParser struct{}
type mountsParser struct{}

func NewMountsParser() Parser {
	return mountsParser{}
}

func NewMountInfoParser() Parser {
	return mountInfoParser{}
}

func NewParser(format MountTabFormat) Parser {
	switch format {
	case MntFmtMountInfo:
		return NewMountInfoParser()
	default:
		return NewMountsParser()
	}
}

// parse one line of mountinfo file
func (p mountInfoParser) Parse(line string) (*Filesystem, error) {
	return nil, errors.New("not implemented")
}

// parse one line of mounts/{fs, m}tab file
func (p mountsParser) Parse(line string) (*Filesystem, error) {
	fs := NewFilesystem()
	tokens := strings.Fields(line)
	// [1] source
	p.parseSourceToken(fs, tokens[0])

	// [2] target
	fs.SetTarget(tokens[1])

	// [3] FS type
	fs.SetFsType(tokens[2])

	// TODO: parse further columns

	return fs, nil
}

func (p mountsParser) parseSourceToken(fs *Filesystem, token string) {
	tag, val, err := parseTagString(token)
	if err != nil && isValidTagName(tag) {
		fs.SetTag(tag, val)
	}
	fs.SetSource(token)
}

func parseTagString(token string) (string, string, error) {
	tv := strings.SplitN(token, "=", 2)
	if len(tv) < 2 {
		return "", "", ErrInvalidToken
	}
	return tv[0], tv[1], nil
}

func isValidTagName(tag string) bool {
	return (tag == "ID" ||
		tag == "UUID" ||
		tag == "LABEL" ||
		tag == "PARTUUID" ||
		tag == "PARTLABEL")
}
