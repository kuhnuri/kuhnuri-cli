package main

import (
	"net/url"
	"path/filepath"
)

func toUrl(in string) (*url.URL, error) {
	uri, err := url.Parse(in)
	if err != nil {
		abs, err := filepath.Abs(in)
		if err != nil {
			return nil, err
		}
		return &url.URL{
			"file",
			"",
			nil,
			"",
			"",
			filepath.ToSlash(abs),
			false,
			"",
			"",
		}, nil
	} else if uri.IsAbs() {
		return uri, nil
	} else {
		abs, err := filepath.Abs(uri.Path)
		if err != nil {
			return nil, err
		}
		return &url.URL{
			"file",
			uri.Opaque,
			uri.User,
			uri.Host,
			abs,
			"",
			uri.ForceQuery,
			uri.RawQuery,
			uri.Fragment,
		}, nil
	}
}
