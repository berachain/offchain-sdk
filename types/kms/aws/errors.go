package aws

import "errors"

var ErrPublicKeyReconstruction = errors.New("can not reconstruct public key from sig")
