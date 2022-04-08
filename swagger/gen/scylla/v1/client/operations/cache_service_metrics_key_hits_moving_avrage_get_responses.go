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

// CacheServiceMetricsKeyHitsMovingAvrageGetReader is a Reader for the CacheServiceMetricsKeyHitsMovingAvrageGet structure.
type CacheServiceMetricsKeyHitsMovingAvrageGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CacheServiceMetricsKeyHitsMovingAvrageGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewCacheServiceMetricsKeyHitsMovingAvrageGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewCacheServiceMetricsKeyHitsMovingAvrageGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCacheServiceMetricsKeyHitsMovingAvrageGetOK creates a CacheServiceMetricsKeyHitsMovingAvrageGetOK with default headers values
func NewCacheServiceMetricsKeyHitsMovingAvrageGetOK() *CacheServiceMetricsKeyHitsMovingAvrageGetOK {
	return &CacheServiceMetricsKeyHitsMovingAvrageGetOK{}
}

/*CacheServiceMetricsKeyHitsMovingAvrageGetOK handles this case with default header values.

CacheServiceMetricsKeyHitsMovingAvrageGetOK cache service metrics key hits moving avrage get o k
*/
type CacheServiceMetricsKeyHitsMovingAvrageGetOK struct {
	Payload *models.RateMovingAverage
}

func (o *CacheServiceMetricsKeyHitsMovingAvrageGetOK) GetPayload() *models.RateMovingAverage {
	return o.Payload
}

func (o *CacheServiceMetricsKeyHitsMovingAvrageGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RateMovingAverage)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCacheServiceMetricsKeyHitsMovingAvrageGetDefault creates a CacheServiceMetricsKeyHitsMovingAvrageGetDefault with default headers values
func NewCacheServiceMetricsKeyHitsMovingAvrageGetDefault(code int) *CacheServiceMetricsKeyHitsMovingAvrageGetDefault {
	return &CacheServiceMetricsKeyHitsMovingAvrageGetDefault{
		_statusCode: code,
	}
}

/*CacheServiceMetricsKeyHitsMovingAvrageGetDefault handles this case with default header values.

internal server error
*/
type CacheServiceMetricsKeyHitsMovingAvrageGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the cache service metrics key hits moving avrage get default response
func (o *CacheServiceMetricsKeyHitsMovingAvrageGetDefault) Code() int {
	return o._statusCode
}

func (o *CacheServiceMetricsKeyHitsMovingAvrageGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *CacheServiceMetricsKeyHitsMovingAvrageGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *CacheServiceMetricsKeyHitsMovingAvrageGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
