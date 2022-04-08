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

// StorageServiceSlowQueryGetReader is a Reader for the StorageServiceSlowQueryGet structure.
type StorageServiceSlowQueryGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StorageServiceSlowQueryGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStorageServiceSlowQueryGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStorageServiceSlowQueryGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStorageServiceSlowQueryGetOK creates a StorageServiceSlowQueryGetOK with default headers values
func NewStorageServiceSlowQueryGetOK() *StorageServiceSlowQueryGetOK {
	return &StorageServiceSlowQueryGetOK{}
}

/*StorageServiceSlowQueryGetOK handles this case with default header values.

StorageServiceSlowQueryGetOK storage service slow query get o k
*/
type StorageServiceSlowQueryGetOK struct {
	Payload *models.SlowQueryInfo
}

func (o *StorageServiceSlowQueryGetOK) GetPayload() *models.SlowQueryInfo {
	return o.Payload
}

func (o *StorageServiceSlowQueryGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.SlowQueryInfo)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewStorageServiceSlowQueryGetDefault creates a StorageServiceSlowQueryGetDefault with default headers values
func NewStorageServiceSlowQueryGetDefault(code int) *StorageServiceSlowQueryGetDefault {
	return &StorageServiceSlowQueryGetDefault{
		_statusCode: code,
	}
}

/*StorageServiceSlowQueryGetDefault handles this case with default header values.

internal server error
*/
type StorageServiceSlowQueryGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the storage service slow query get default response
func (o *StorageServiceSlowQueryGetDefault) Code() int {
	return o._statusCode
}

func (o *StorageServiceSlowQueryGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *StorageServiceSlowQueryGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *StorageServiceSlowQueryGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
