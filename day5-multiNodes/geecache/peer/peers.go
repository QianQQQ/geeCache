package peer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PeerPicker interface {
	PickPeer(key string) (peer CacheGetter, ok bool)
}

type CacheGetter interface {
	Get(group, key string) ([]byte, error)
}

type HTTPGetter struct {
	BaseURL string
}

func (g *HTTPGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", g.BaseURL, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("server return: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}
