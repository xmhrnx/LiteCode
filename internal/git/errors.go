package git

import "errors"

var errNoRemote = errors.New("the repository has no remote to push to")
