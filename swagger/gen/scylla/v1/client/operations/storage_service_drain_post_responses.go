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

// StorageServiceDrainPostReader is a Reader for the StorageServiceDrainPost structure.
type StorageServiceDrainPostReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StorageServiceDrainPostReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStorageServiceDrainPostOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStorageServiceDrainPostDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStorageServiceDrainPostOK creates a StorageServiceDrainPostOK with default headers values
func NewStorageServiceDrainPostOK() *StorageServiceDrainPostOK {
	return &StorageServiceDrainPostOK{}
}

/*StorageServiceDrainPostOK handles this case with default header values.

StorageServiceDrainPostOK storage service drain post o k
*/
type StorageServiceDrainPostOK struct {
}

func (o *StorageServiceDrainPostOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewStorageServiceDrainPostDefault creates a StorageServiceDrainPostDefault with default headers values
func NewStorageServiceDrainPostDefault(code int) *StorageServiceDrainPostDefault {
	return &StorageServiceDrainPostDefault{
		_statusCode: code,
	}
}

/*StorageServiceDrainPostDefault handles this case with default header values.

internal server error
*/
type StorageServiceDrainPostDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the storage service drain post default response
func (o *StorageServiceDrainPostDefault) Code() int {
	return o._statusCode
}

func (o *StorageServiceDrainPostDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *StorageServiceDrainPostDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *StorageServiceDrainPostDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
