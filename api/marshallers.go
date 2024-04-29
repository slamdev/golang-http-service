package api

import (
	"encoding/json"
	"errors"
)

// ProblemDetail has Detail field that is mapped to error in go generated code
// error type has no json marshalling support so we define custom marshaller and unmarshaller
// to convert error to string and vice versa

func (r NotFoundApplicationProblemPlusJSONResponse) MarshalJSON() ([]byte, error) {
	p := ProblemDetail(r)
	return p.MarshalJSON()
}

func (r BadRequestApplicationProblemPlusJSONResponse) MarshalJSON() ([]byte, error) {
	p := ProblemDetail(r)
	return p.MarshalJSON()
}

func (p ProblemDetail) MarshalJSON() ([]byte, error) {
	type Alias ProblemDetail
	var errStr string
	if p.Detail != nil {
		errStr = p.Detail.Error()
	}
	return json.Marshal(&struct {
		Detail string `json:"detail"`
		*Alias
	}{
		Detail: errStr,
		Alias:  (*Alias)(&p),
	})
}

func (p *ProblemDetail) UnmarshalJSON(data []byte) error {
	type Alias ProblemDetail
	var res struct {
		Detail string `json:"detail"`
		*Alias
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	p.Type = res.Type
	p.Title = res.Title
	p.Status = res.Status
	p.Instance = res.Instance
	p.TraceId = res.TraceId
	p.Detail = errors.New(res.Detail)
	return nil
}
