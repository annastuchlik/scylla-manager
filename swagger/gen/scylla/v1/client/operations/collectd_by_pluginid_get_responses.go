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

// CollectdByPluginidGetReader is a Reader for the CollectdByPluginidGet structure.
type CollectdByPluginidGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CollectdByPluginidGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewCollectdByPluginidGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewCollectdByPluginidGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCollectdByPluginidGetOK creates a CollectdByPluginidGetOK with default headers values
func NewCollectdByPluginidGetOK() *CollectdByPluginidGetOK {
	return &CollectdByPluginidGetOK{}
}

/*CollectdByPluginidGetOK handles this case with default header values.

CollectdByPluginidGetOK collectd by pluginid get o k
*/
type CollectdByPluginidGetOK struct {
	Payload []*models.CollectdValue
}

func (o *CollectdByPluginidGetOK) GetPayload() []*models.CollectdValue {
	return o.Payload
}

func (o *CollectdByPluginidGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCollectdByPluginidGetDefault creates a CollectdByPluginidGetDefault with default headers values
func NewCollectdByPluginidGetDefault(code int) *CollectdByPluginidGetDefault {
	return &CollectdByPluginidGetDefault{
		_statusCode: code,
	}
}

/*CollectdByPluginidGetDefault handles this case with default header values.

internal server error
*/
type CollectdByPluginidGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the collectd by pluginid get default response
func (o *CollectdByPluginidGetDefault) Code() int {
	return o._statusCode
}

func (o *CollectdByPluginidGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *CollectdByPluginidGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *CollectdByPluginidGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
