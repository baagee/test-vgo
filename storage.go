package main

type Storage interface {
	Shorten(url string, expire int64) (string, error)
	ShortenInfo(shorten string) (UrlDetail, error)
	UnShorten(shorten string) (string, error)
}
