// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"net/url"
	"strconv"
)

// GetSubnetsRequest describes the parameters to request a list of subnets.
type GetSubnetsRequest struct {
	Page    int
	PerPage int
	Free    bool
}

// ApplyToURL modifies the given url to include query string parameters for the request.
func (request *GetSubnetsRequest) ApplyToURL(u *url.URL) {
	q := u.Query()
	q.Add("page", strconv.Itoa(request.Page))
	q.Add("per_page", strconv.Itoa(request.PerPage))
	if request.Free {
		q.Add("show_free", "true")
	}
	u.RawQuery = q.Encode()
}
