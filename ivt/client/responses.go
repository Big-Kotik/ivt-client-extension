package client

import (
	"inv-client-extension/ivt/types"
	"net/http"
)

func (user *User) SendResponses(responses types.ResponsesWrapper) {
	for _, res := range responses.Data {
		val, ok := user.idToRequest.LoadAndDelete(res.ID.String())
		if !ok {
			continue
		}
		req, ok := val.(types.Requests)
		if !ok {
			continue
		}
		req.Write(res.Body)
		h := http.Header{}
		for k, vals := range res.Headers {
			for _, val := range vals {
				h.Add(k, val)
			}
		}
		req.WriteHeaders(h)
		req.Complete()
	}
}
