package client

import (
	"inv-client-extension/ivt/types"
	"log"
	"net/http"
)

func (user *User) SendResponses(responses *types.ResponsesWrapper) {
	for _, res := range responses.Data {
		user.idToRequest.Range(func(key, value interface{}) bool {
			return true
		})

		val, ok := user.idToRequest.LoadAndDelete(res.ID.String())
		if !ok {
			log.Printf("not ok")
			continue
		}
		req, ok := val.(types.Requests)
		if !ok {
			continue
		}

		h := http.Header{}
		for k, vals := range res.Headers {
			for _, val := range vals {
				h.Add(k, val)
			}
		}
		req.WriteHeaders(h)
		req.Write(res.Body)
		req.Complete()
	}
}
