// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/scylladb/scylla-manager/v3/swagger/gen/agent/models"
)

// CoreStatsDeleteReader is a Reader for the CoreStatsDelete structure.
type CoreStatsDeleteReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CoreStatsDeleteReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewCoreStatsDeleteOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewCoreStatsDeleteDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCoreStatsDeleteOK creates a CoreStatsDeleteOK with default headers values
func NewCoreStatsDeleteOK() *CoreStatsDeleteOK {
	return &CoreStatsDeleteOK{}
}

/*CoreStatsDeleteOK handles this case with default header values.

Empty object
*/
type CoreStatsDeleteOK struct {
	Payload interface{}
	JobID   int64
}

func (o *CoreStatsDeleteOK) GetPayload() interface{} {
	return o.Payload
}

func (o *CoreStatsDeleteOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	if jobIDHeader := response.GetHeader("x-rclone-jobid"); jobIDHeader != "" {
		jobID, err := strconv.ParseInt(jobIDHeader, 10, 64)
		if err != nil {
			return err
		}

		o.JobID = jobID
	}
	return nil
}

// NewCoreStatsDeleteDefault creates a CoreStatsDeleteDefault with default headers values
func NewCoreStatsDeleteDefault(code int) *CoreStatsDeleteDefault {
	return &CoreStatsDeleteDefault{
		_statusCode: code,
	}
}

/*CoreStatsDeleteDefault handles this case with default header values.

Server error
*/
type CoreStatsDeleteDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
	JobID   int64
}

// Code gets the status code for the core stats delete default response
func (o *CoreStatsDeleteDefault) Code() int {
	return o._statusCode
}

func (o *CoreStatsDeleteDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CoreStatsDeleteDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	if jobIDHeader := response.GetHeader("x-rclone-jobid"); jobIDHeader != "" {
		jobID, err := strconv.ParseInt(jobIDHeader, 10, 64)
		if err != nil {
			return err
		}

		o.JobID = jobID
	}
	return nil
}

func (o *CoreStatsDeleteDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
