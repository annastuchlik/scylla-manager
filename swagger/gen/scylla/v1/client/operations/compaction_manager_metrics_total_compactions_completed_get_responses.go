// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/scylladb/scylla-manager/v3/swagger/gen/scylla/v1/models"
)

// CompactionManagerMetricsTotalCompactionsCompletedGetReader is a Reader for the CompactionManagerMetricsTotalCompactionsCompletedGet structure.
type CompactionManagerMetricsTotalCompactionsCompletedGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CompactionManagerMetricsTotalCompactionsCompletedGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewCompactionManagerMetricsTotalCompactionsCompletedGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewCompactionManagerMetricsTotalCompactionsCompletedGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCompactionManagerMetricsTotalCompactionsCompletedGetOK creates a CompactionManagerMetricsTotalCompactionsCompletedGetOK with default headers values
func NewCompactionManagerMetricsTotalCompactionsCompletedGetOK() *CompactionManagerMetricsTotalCompactionsCompletedGetOK {
	return &CompactionManagerMetricsTotalCompactionsCompletedGetOK{}
}

/*CompactionManagerMetricsTotalCompactionsCompletedGetOK handles this case with default header values.

CompactionManagerMetricsTotalCompactionsCompletedGetOK compaction manager metrics total compactions completed get o k
*/
type CompactionManagerMetricsTotalCompactionsCompletedGetOK struct {
	Payload interface{}
}

func (o *CompactionManagerMetricsTotalCompactionsCompletedGetOK) GetPayload() interface{} {
	return o.Payload
}

func (o *CompactionManagerMetricsTotalCompactionsCompletedGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCompactionManagerMetricsTotalCompactionsCompletedGetDefault creates a CompactionManagerMetricsTotalCompactionsCompletedGetDefault with default headers values
func NewCompactionManagerMetricsTotalCompactionsCompletedGetDefault(code int) *CompactionManagerMetricsTotalCompactionsCompletedGetDefault {
	return &CompactionManagerMetricsTotalCompactionsCompletedGetDefault{
		_statusCode: code,
	}
}

/*CompactionManagerMetricsTotalCompactionsCompletedGetDefault handles this case with default header values.

internal server error
*/
type CompactionManagerMetricsTotalCompactionsCompletedGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the compaction manager metrics total compactions completed get default response
func (o *CompactionManagerMetricsTotalCompactionsCompletedGetDefault) Code() int {
	return o._statusCode
}

func (o *CompactionManagerMetricsTotalCompactionsCompletedGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *CompactionManagerMetricsTotalCompactionsCompletedGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *CompactionManagerMetricsTotalCompactionsCompletedGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
