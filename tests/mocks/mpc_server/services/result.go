package services

import "github.com/pkg/errors"

type Provider struct{}

func (o *Provider) Result(in *ResultInput) (*ResultOutput, error) {
	out := new(ResultOutput)
	lastKeygenReq := storer.GetKeygenRequestModel(in.RequestId)
	if lastKeygenReq != nil {
		out.RequestId = in.RequestId
		out.RequestType = lastKeygenReq.reqType
		out.RequestStatus = lastKeygenReq.status
		out.Result = lastKeygenReq.result
		return out, nil
	}

	lastSignReq := storer.GetSignRequestModel(in.RequestId)
	if lastSignReq == nil {
		return nil, errors.Errorf("request id %q not exists", in.RequestId)
	}
	out.RequestId = in.RequestId
	out.RequestType = lastSignReq.reqType
	out.RequestStatus = lastSignReq.status
	out.Result = lastSignReq.result
	return out, nil
}
