/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package common

import (
	"errors"
	"fmt"
)

type ValidationResult struct {
	Result      Result
	CheckID     ValidationCheck
	ComponentID string
	Description string
}

func NewValidationResult(result Result, id ValidationCheck, componentId string, description string) ValidationResult {
	return ValidationResult{
		Result:      result,
		CheckID:     id,
		ComponentID: componentId,
		Description: description,
	}
}

type ValidationCheck string

const (
	IPRangeConflictCheck ValidationCheck = "ip-range-conflict"
	SLSSchemaCheck       ValidationCheck = "sls-schema-validation"
)

type Result string

const (
	Fail    Result = "fail"
	Warning Result = "warning"
	Pass    Result = "pass"
)

type ValidationResults struct {
	results []ValidationResult
}

func NewValidationResults() *ValidationResults {
	var v ValidationResults
	v.results = make([]ValidationResult, 0)
	return &v
}

func (v *ValidationResults) GetResults() []ValidationResult {
	return v.results
}

func (v *ValidationResults) ToError() error {
	return AllError(v.results)
}

func (v *ValidationResults) Add(results ...ValidationResult) {
	v.results = append(v.results, results...)
}

func (v *ValidationResults) Fail(id ValidationCheck, componentId string, description string) {
	result := NewValidationResult(Fail, id, componentId, description)
	v.results = append(v.results, result)
}

func (v *ValidationResults) Warning(id ValidationCheck, componentId string, description string) {
	result := NewValidationResult(Warning, id, componentId, description)
	v.results = append(v.results, result)
}

func (v *ValidationResults) Pass(id ValidationCheck, componentId string, description string) {
	result := NewValidationResult(Pass, id, componentId, description)
	v.results = append(v.results, result)
}

func FailureResult(id ValidationCheck, componentId string, description string) ValidationResult {
	return NewValidationResult(Fail, id, componentId, description)
}

func WarningResult(id ValidationCheck, componentId string, description string) ValidationResult {
	return NewValidationResult(Warning, id, componentId, description)
}

func PassResult(id ValidationCheck, componentId string, description string) ValidationResult {
	return NewValidationResult(Pass, id, componentId, description)
}

func AllError(results []ValidationResult) error {
	var allError error
	for _, result := range results {
		if result.Result == Fail {
			e := fmt.Errorf("%s: %s", result.ComponentID, result.Description)
			allError = errors.Join(allError, e)
		}
	}
	return allError
}
