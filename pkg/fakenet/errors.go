package fakenet

import "github.com/k42-software/go-multierror/v2"

// Small performance optimisation
func appendErr(err error, errs ...error) error {
	switch len(errs) {
	case 0:
		return err

	case 1:
		if err == nil {
			return errs[0]
		} else if errs[0] == nil {
			return err
		}
		fallthrough

	default:
		return multierror.Append(err, errs...)
	}
}
