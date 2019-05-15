// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// Job job
//
// Status information about the job
// swagger:model Job
type Job struct {

	// Time in seconds that the job ran for
	Duration int64 `json:"duration,omitempty"`

	// Time the job finished (eg 2018-10-26T18:50:20.528746884+01:00)
	EndTime string `json:"endTime,omitempty"`

	// Error from the job or empty string for no error
	Error string `json:"error,omitempty"`

	// Job has finished execution
	Finished bool `json:"finished,omitempty"`

	// ID of the job
	ID int64 `json:"id,omitempty"`

	// Output of the job as would have been returned if called synchronously
	Output interface{} `json:"output,omitempty"`

	// Time the job started (eg 2018-10-26T18:50:20.528746884+01:00)
	StartTime string `json:"startTime,omitempty"`

	// True for success false otherwise
	Success bool `json:"success,omitempty"`
}

// Validate validates this job
func (m *Job) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Job) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Job) UnmarshalBinary(b []byte) error {
	var res Job
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
