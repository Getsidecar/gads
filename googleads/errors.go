package v201809

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type baseError struct {
	code    string
	origErr error
}

func (b baseError) Error() string {
	return b.origErr.Error()
}

func (b baseError) String() string {
	return b.Error()
}

func (b baseError) Code() string {
	return b.code
}

func (b baseError) OrigErr() error {
	return b.origErr
}

type Error interface {
	// Satisfy the generic error interface.
	error

	// Returns the short phrase depicting the classification of the error.
	Code() string

	// Returns the original error if one was set.  Nil is returned if not set.
	OrigErr() error
}

type OperationError struct {
	Code      int64  `xml:"OperationError>Code"`
	Details   string `xml:"OperationError>Details"`
	ErrorCode string `xml:"OperationError>ErrorCode"`
	Message   string `xml:"OperationError>Message"`
}

type EntityError struct {
	FieldPath   string `xml:"fieldPath"`
	Trigger     string `xml:"trigger"`
	ErrorString string `xml:"errorString"`
	Reason      string `xml:"reason"`
}

type BudgetError struct {
	EntityError
}

type CriterionError struct {
	EntityError
}

type TargetError struct {
	FieldPath   string `xml:"fieldPath"`
	Trigger     string `xml:"trigger"`
	ErrorString string `xml:"errorString"`
	Reason      string `xml:"reason"`
}

type AdGroupServiceError struct {
	FieldPath   string `xml:"fieldPath"`
	Trigger     string `xml:"trigger"`
	ErrorString string `xml:"errorString"`
	Reason      string `xml:"reason"`
}

type NotEmptyError struct {
	FieldPath   string `xml:"fieldPath"`
	Trigger     string `xml:"trigger"`
	ErrorString string `xml:"errorString"`
	Reason      string `xml:"reason"`
}

type AdError struct {
	FieldPath   string `xml:"fieldPath"`
	Trigger     string `xml:"trigger"`
	ErrorString string `xml:"errorString"`
	Reason      string `xml:"reason"`
}

type LabelError struct {
	FieldPath   string `xml:"fieldPath"`
	Trigger     string `xml:"trigger"`
	ErrorString string `xml:"errorString"`
	Reason      string `xml:"reason"`
}

type UrlError struct {
	EntityError
}

// if you exceed the quota given by google
type RateExceededError struct {
	RateName          string `xml:"rateName"`  // For example OperationsByMinute
	RateScope         string `xml:"rateScope"` // ACCOUNT or DEVELOPER
	ErrorString       string `xml:"errorString"`
	Reason            string `xml:"reason"`
	RetryAfterSeconds uint   `xml:"retryAfterSeconds"` // Try again in...
}

type ApiExceptionFault struct {
	Message    string `xml:"message"`
	Type       string `xml:"ApplicationException.Type"`
	ErrorsType string
	Reason     string
	Errors     []interface{} `xml:"errors"`
}

func (aes *ApiExceptionFault) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) (err error) {
	for token, err := dec.Token(); err == nil; token, err = dec.Token() {
		switch start := token.(type) {
		case xml.StartElement:
			switch start.Name.Local {
			case "message":
				if err := dec.DecodeElement(&aes.Message, &start); err != nil {
					return err
				}
			case "ApplicationException.Type":
				if err := dec.DecodeElement(&aes.Type, &start); err != nil {
					return err
				}
			case "errors":
				errorType, _ := findAttr(start.Attr, xml.Name{Space: "http://www.w3.org/2001/XMLSchema-instance", Local: "type"})
				aes.ErrorsType = errorType

				switch errorType {
				case "RateExceededError":
					e := RateExceededError{}
					dec.DecodeElement(&e, &start)
					aes.Errors = append(aes.Errors, e)
					aes.Reason = e.Reason

				case "AuthenticationError", "DatabaseError", "InternalApiError":
					e := EntityError{}
					if err := dec.DecodeElement(&e, &start); err != nil {
						return fmt.Errorf("Unknown error type -> %s", start)
					}
					aes.Errors = append(aes.Errors, e)
					aes.Reason = e.Reason

				default:
					e := EntityError{}
					if err := dec.DecodeElement(&e, &start); err != nil {
						return fmt.Errorf("Unknown error type -> %s", start)
					}
					aes.Errors = append(aes.Errors, e)
				}
			case "reason":
				break
			default:
				return fmt.Errorf("Unknown error field -> %s", start)
			}
		}
	}
	return err
}

type ErrorsType struct {
	ApiExceptionFaults []ApiExceptionFault `xml:"ApiExceptionFault"`
}

func (f ErrorsType) Error() string {
	errors := []string{}
	for _, e := range f.ApiExceptionFaults {
		errors = append(errors, fmt.Sprintf("%s", e.Message))
	}
	return strings.Join(errors, "\n")
}

type Fault struct {
	XMLName     xml.Name   `xml:"Fault"`
	FaultCode   string     `xml:"faultcode"`
	FaultString string     `xml:"faultstring"`
	Errors      ErrorsType `xml:"detail"`
}

func (f Fault) Error() string {
	return f.FaultString + " - " + f.Errors.Error()
}
