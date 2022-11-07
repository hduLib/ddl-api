package client

import (
	"errors"
	"net/http"
)

type WaitingClient struct {
	*http.Client
	pool chan struct{}
}

func (w *WaitingClient) Do(request *http.Request) (*http.Response, error) {
	<-w.pool
	resp, err := w.Client.Do(request)
	w.pool <- struct{}{}
	return resp, err
}

func NewWaitingClient(thread int) *WaitingClient {
	c := WaitingClient{
		Client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				for _, v := range append(via[len(via)-1].Cookies(), req.Response.Cookies()...) {
					ck, err := req.Cookie(v.Name)
					if err != nil {
						if errors.Is(err, http.ErrNoCookie) {
							req.AddCookie(v)
						} else {
							return err
						}
						continue
					}
					ck.Value = v.Value
				}
				return nil
			},
		},
		pool: make(chan struct{}, thread),
	}
	for i := 0; i < thread; i++ {
		c.pool <- struct{}{}
	}
	return &c
}
