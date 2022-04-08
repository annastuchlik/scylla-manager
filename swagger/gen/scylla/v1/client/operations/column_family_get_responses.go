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

// ColumnFamilyGetReader is a Reader for the ColumnFamilyGet structure.
type ColumnFamilyGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ColumnFamilyGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewColumnFamilyGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewColumnFamilyGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewColumnFamilyGetOK creates a ColumnFamilyGetOK with default headers values
func NewColumnFamilyGetOK() *ColumnFamilyGetOK {
	return &ColumnFamilyGetOK{}
}

/*ColumnFamilyGetOK handles this case with default header values.

ColumnFamilyGetOK column family get o k
*/
type ColumnFamilyGetOK struct {
	Payload []*models.ColumnFamilyInfo
}

func (o *ColumnFamilyGetOK) GetPayload() []*models.ColumnFamilyInfo {
	return o.Payload
}

func (o *ColumnFamilyGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewColumnFamilyGetDefault creates a ColumnFamilyGetDefault with default headers values
func NewColumnFamilyGetDefault(code int) *ColumnFamilyGetDefault {
	return &ColumnFamilyGetDefault{
		_statusCode: code,
	}
}

/*ColumnFamilyGetDefault handles this case with default header values.

internal server error
*/
type ColumnFamilyGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the column family get default response
func (o *ColumnFamilyGetDefault) Code() int {
	return o._statusCode
}

func (o *ColumnFamilyGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *ColumnFamilyGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *ColumnFamilyGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
