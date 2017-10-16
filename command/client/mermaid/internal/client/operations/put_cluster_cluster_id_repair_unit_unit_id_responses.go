// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/scylladb/mermaid/command/client/mermaid/internal/models"
)

// PutClusterClusterIDRepairUnitUnitIDReader is a Reader for the PutClusterClusterIDRepairUnitUnitID structure.
type PutClusterClusterIDRepairUnitUnitIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PutClusterClusterIDRepairUnitUnitIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewPutClusterClusterIDRepairUnitUnitIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		body := response.Body()
		defer body.Close()

		b, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		json.Indent(buf, b, "", "  ")

		return nil, runtime.NewAPIError("API error", "\n"+buf.String(), response.Code())
	}
}

// NewPutClusterClusterIDRepairUnitUnitIDOK creates a PutClusterClusterIDRepairUnitUnitIDOK with default headers values
func NewPutClusterClusterIDRepairUnitUnitIDOK() *PutClusterClusterIDRepairUnitUnitIDOK {
	return &PutClusterClusterIDRepairUnitUnitIDOK{}
}

/*PutClusterClusterIDRepairUnitUnitIDOK handles this case with default header values.

updated unit fields
*/
type PutClusterClusterIDRepairUnitUnitIDOK struct {
	Payload *models.RepairUnit
}

func (o *PutClusterClusterIDRepairUnitUnitIDOK) Error() string {
	return fmt.Sprintf("[PUT /cluster/{cluster_id}/repair/unit/{unit_id}][%d] putClusterClusterIdRepairUnitUnitIdOK  %+v", 200, o.Payload)
}

func (o *PutClusterClusterIDRepairUnitUnitIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RepairUnit)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
