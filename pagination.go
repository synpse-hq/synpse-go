package synpse

import (
	"net/http"
	"net/url"
	"strconv"
)

// PaginationOptions specifies pagination options for the API client, it's used together with the
// endpoints that support pagination
type PaginationOptions struct {
	PageToken string
	PageSize  int
}

type Pagination struct {
	NextPageToken     string
	PreviousPageToken string
	PageSize          int
	TotalItems        int
}

// Pagination headers
var (
	HeaderNextPageToken     = "Next-Page-Token"
	HeaderPreviousPageToken = "Previous-Page-Token"
	HeaderPageSize          = "Page-Size"
	HeaderTotalItems        = "Total-Items"
)

// Common pagination query args
var (
	PaginationQueryPageToken = "pageToken"
	PaginationPageSize       = "pageSize"
)

const (
	DefaultPageSize = 100
	MaxPageSize     = 500
)

func setPagination(urlValues url.Values, opts *PaginationOptions) {
	if opts.PageSize > 0 {
		urlValues.Set(PaginationPageSize, strconv.Itoa(opts.PageSize))
	}
	if opts.PageToken != "" {
		urlValues.Set(PaginationQueryPageToken, opts.PageToken)
	}
}

func getPagination(respHeader http.Header) Pagination {
	p := Pagination{
		NextPageToken:     respHeader.Get(HeaderNextPageToken),
		PreviousPageToken: respHeader.Get(HeaderPreviousPageToken),
	}

	pageSize, err := strconv.Atoi(respHeader.Get(HeaderPageSize))
	if err == nil {
		p.PageSize = pageSize
	}

	totalItems, err := strconv.Atoi(respHeader.Get(HeaderTotalItems))
	if err == nil {
		p.TotalItems = totalItems
	}
	return p
}
