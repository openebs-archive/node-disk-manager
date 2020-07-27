/*
Copyright 2018 The OpenEBS Authors.

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

package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckTruthy(t *testing.T) {
	tests := map[string]struct {
		checkString string
		expected    bool
	}{
		"1 should return true":      {checkString: "1", expected: true},
		"yes should return true":    {checkString: "Yes", expected: true},
		"ok should return true":     {checkString: "Ok", expected: true},
		"true should return true":   {checkString: "True", expected: true},
		"0 should return false":     {checkString: "0", expected: false},
		"no should return false":    {checkString: "No", expected: false},
		"false should return false": {checkString: "False", expected: false},
		"blank should return false": {checkString: "", expected: false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, CheckTruthy(test.checkString))
		})
	}
}

func TestCheckFalsy(t *testing.T) {
	tests := map[string]struct {
		checkString string
		expected    bool
	}{
		"1 should return false":    {checkString: "1", expected: false},
		"yes should return false":  {checkString: "Yes", expected: false},
		"ok should return false":   {checkString: "Ok", expected: false},
		"true should return false": {checkString: "True", expected: false},
		"0 should return true":     {checkString: "0", expected: true},
		"no should return true":    {checkString: "No", expected: true},
		"false should return true": {checkString: "False", expected: true},
		"blank should return true": {checkString: "", expected: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, CheckFalsy(test.checkString))
		})
	}
}

func TestStringToInt32(t *testing.T) {
	tests := map[string]struct {
		numberString string
		expected     int32
	}{
		"string 6 should return int 6":                {numberString: "6", expected: 6},
		"string 18 should return int 18":              {numberString: "18", expected: 18},
		"string 32 should return int 32":              {numberString: "32", expected: 32},
		" should return nil pinter and one error":     {numberString: "", expected: 0},
		"test should return nil pinter and one error": {numberString: "test", expected: 0},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualInt, err := StringToInt32(test.numberString)
			if err != nil {
				return
			}
			assert.Equal(t, test.expected, *actualInt)
		})
	}
}

func TestStrToInt32(t *testing.T) {
	tests := map[string]struct {
		numberString string
		expected     int32
	}{
		"string 6 should return int 6":   {numberString: "6", expected: 6},
		"string 18 should return int 18": {numberString: "18", expected: 18},
		"string 32 should return int 32": {numberString: "32", expected: 32},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualInt := StrToInt32(test.numberString)
			assert.Equal(t, test.expected, *actualInt)
		})
	}
}

func TestHash(t *testing.T) {
	tests := map[string]struct {
		hashString string
		expected   string
	}{
		"check hash for string1": {hashString: "This is one string", expected: "6192a12ec601c65b8375743eb66167ab"},
		"check hash for string2": {hashString: "This is one string", expected: "6192a12ec601c65b8375743eb66167ab"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Hash(test.hashString))
		})
	}
}

func TestCheckErr(t *testing.T) {
	errStr := "This is a test string"
	err := errors.New(errStr)
	var msgChannel = make(chan string)
	handlerFunc := func(str string) {
		msgChannel <- str
	}
	go func() {
		msg := <-msgChannel
		assert.Equal(t, errStr, msg)
	}()
	CheckErr(err, handlerFunc)
	CheckErr(nil, handlerFunc)
}

func TestStateStatus(t *testing.T) {
	tests := map[string]struct {
		status string
		state  bool
	}{
		"status is enable for true state":   {state: true, status: "enable"},
		"status is disable for false state": {state: false, status: "disable"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.status, StateStatus(test.state))
		})
	}
}

func TestConvBytesToGigabytes(t *testing.T) {

	var oneGBinBytes uint64 = 1073741824
	var threeDotFiveinBytes = float64(oneGBinBytes) * 3.5
	var testCases = []struct {
		testName    string
		sizeinBytes uint64
		exp         float64
	}{
		{
			testName:    "Converting bytes to 1GB",
			sizeinBytes: oneGBinBytes,
			exp:         1,
		},
		{
			testName:    "Converting bytes to 2GB",
			sizeinBytes: 2 * oneGBinBytes,
			exp:         2,
		},
		{
			testName:    "Converting bytes to 3.5GB to check decimal values",
			sizeinBytes: uint64(threeDotFiveinBytes),
			exp:         3.5,
		},
		{
			testName:    "Converting bytes to 1PetaBytes",
			sizeinBytes: 1024 * 1024 * oneGBinBytes,
			exp:         1024 * 1024,
		},
	}

	for _, tc := range testCases {
		res := ConvBytesToGigabytes(tc.sizeinBytes)
		if !assert.Equal(t, res, tc.exp) {
			t.Errorf(" Test failed : %v , expected : %v  , got : %v", tc.testName, tc.exp, res)
		}

	}
}
