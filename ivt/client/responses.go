package client

import (
	"inv-client-extension/ivt/types"
	"log"
	"net/http"
)

func (user *User) SendResponses(responses *types.ResponsesWrapper) {
	log.Printf("res: %v", responses.Data[0])

	for _, res := range responses.Data {

		log.Printf("res2: %v", res)
		log.Print(res.ID.String())

		user.idToRequest.Range(func(key, value interface{}) bool {
			log.Printf("key: %v", key)
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

		n, err := req.Write(res.Body)
		log.Print(n, err)

		req.Complete()
	}
}
