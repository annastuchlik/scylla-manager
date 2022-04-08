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

// StorageServiceTokensByEndpointGetReader is a Reader for the StorageServiceTokensByEndpointGet structure.
type StorageServiceTokensByEndpointGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StorageServiceTokensByEndpointGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStorageServiceTokensByEndpointGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStorageServiceTokensByEndpointGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStorageServiceTokensByEndpointGetOK creates a StorageServiceTokensByEndpointGetOK with default headers values
func NewStorageServiceTokensByEndpointGetOK() *StorageServiceTokensByEndpointGetOK {
	return &StorageServiceTokensByEndpointGetOK{}
}

/*StorageServiceTokensByEndpointGetOK handles this case with default header values.

StorageServiceTokensByEndpointGetOK storage service tokens by endpoint get o k
*/
type StorageServiceTokensByEndpointGetOK struct {
	Payload []string
}

func (o *StorageServiceTokensByEndpointGetOK) GetPayload() []string {
	return o.Payload
}

func (o *StorageServiceTokensByEndpointGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewStorageServiceTokensByEndpointGetDefault creates a StorageServiceTokensByEndpointGetDefault with default headers values
func NewStorageServiceTokensByEndpointGetDefault(code int) *StorageServiceTokensByEndpointGetDefault {
	return &StorageServiceTokensByEndpointGetDefault{
		_statusCode: code,
	}
}

/*StorageServiceTokensByEndpointGetDefault handles this case with default header values.

internal server error
*/
type StorageServiceTokensByEndpointGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the storage service tokens by endpoint get default response
func (o *StorageServiceTokensByEndpointGetDefault) Code() int {
	return o._statusCode
}

func (o *StorageServiceTokensByEndpointGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *StorageServiceTokensByEndpointGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *StorageServiceTokensByEndpointGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
