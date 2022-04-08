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

// ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetReader is a Reader for the ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGet structure.
type ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK creates a ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK with default headers values
func NewColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK() *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK {
	return &ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK{}
}

/*ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK handles this case with default header values.

ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK column family metrics bloom filter disk space used by name get o k
*/
type ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK struct {
	Payload interface{}
}

func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK) GetPayload() interface{} {
	return o.Payload
}

func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault creates a ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault with default headers values
func NewColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault(code int) *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault {
	return &ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault{
		_statusCode: code,
	}
}

/*ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault handles this case with default header values.

internal server error
*/
type ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the column family metrics bloom filter disk space used by name get default response
func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault) Code() int {
	return o._statusCode
}

func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *ColumnFamilyMetricsBloomFilterDiskSpaceUsedByNameGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
