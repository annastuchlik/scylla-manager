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

// StorageProxyCounterWriteRPCTimeoutGetReader is a Reader for the StorageProxyCounterWriteRPCTimeoutGet structure.
type StorageProxyCounterWriteRPCTimeoutGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StorageProxyCounterWriteRPCTimeoutGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStorageProxyCounterWriteRPCTimeoutGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStorageProxyCounterWriteRPCTimeoutGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStorageProxyCounterWriteRPCTimeoutGetOK creates a StorageProxyCounterWriteRPCTimeoutGetOK with default headers values
func NewStorageProxyCounterWriteRPCTimeoutGetOK() *StorageProxyCounterWriteRPCTimeoutGetOK {
	return &StorageProxyCounterWriteRPCTimeoutGetOK{}
}

/*StorageProxyCounterWriteRPCTimeoutGetOK handles this case with default header values.

StorageProxyCounterWriteRPCTimeoutGetOK storage proxy counter write Rpc timeout get o k
*/
type StorageProxyCounterWriteRPCTimeoutGetOK struct {
	Payload interface{}
}

func (o *StorageProxyCounterWriteRPCTimeoutGetOK) GetPayload() interface{} {
	return o.Payload
}

func (o *StorageProxyCounterWriteRPCTimeoutGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewStorageProxyCounterWriteRPCTimeoutGetDefault creates a StorageProxyCounterWriteRPCTimeoutGetDefault with default headers values
func NewStorageProxyCounterWriteRPCTimeoutGetDefault(code int) *StorageProxyCounterWriteRPCTimeoutGetDefault {
	return &StorageProxyCounterWriteRPCTimeoutGetDefault{
		_statusCode: code,
	}
}

/*StorageProxyCounterWriteRPCTimeoutGetDefault handles this case with default header values.

internal server error
*/
type StorageProxyCounterWriteRPCTimeoutGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the storage proxy counter write Rpc timeout get default response
func (o *StorageProxyCounterWriteRPCTimeoutGetDefault) Code() int {
	return o._statusCode
}

func (o *StorageProxyCounterWriteRPCTimeoutGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *StorageProxyCounterWriteRPCTimeoutGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *StorageProxyCounterWriteRPCTimeoutGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
