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

// HintedHandoffHintsGetReader is a Reader for the HintedHandoffHintsGet structure.
type HintedHandoffHintsGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *HintedHandoffHintsGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewHintedHandoffHintsGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewHintedHandoffHintsGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewHintedHandoffHintsGetOK creates a HintedHandoffHintsGetOK with default headers values
func NewHintedHandoffHintsGetOK() *HintedHandoffHintsGetOK {
	return &HintedHandoffHintsGetOK{}
}

/*HintedHandoffHintsGetOK handles this case with default header values.

HintedHandoffHintsGetOK hinted handoff hints get o k
*/
type HintedHandoffHintsGetOK struct {
	Payload []string
}

func (o *HintedHandoffHintsGetOK) GetPayload() []string {
	return o.Payload
}

func (o *HintedHandoffHintsGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewHintedHandoffHintsGetDefault creates a HintedHandoffHintsGetDefault with default headers values
func NewHintedHandoffHintsGetDefault(code int) *HintedHandoffHintsGetDefault {
	return &HintedHandoffHintsGetDefault{
		_statusCode: code,
	}
}

/*HintedHandoffHintsGetDefault handles this case with default header values.

internal server error
*/
type HintedHandoffHintsGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the hinted handoff hints get default response
func (o *HintedHandoffHintsGetDefault) Code() int {
	return o._statusCode
}

func (o *HintedHandoffHintsGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *HintedHandoffHintsGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *HintedHandoffHintsGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
