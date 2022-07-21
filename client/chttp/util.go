package chttp

func wrapNoRetryErr(err error) error {
	if err != nil {
		err = &noRetryErr{err: err}
	}
	return err
}

type noRetryErr struct {
	err error
}

func (e *noRetryErr) Error() string {
	return e.err.Error()
}

func unwrapNoRetryErr(err error) error {
	if e, ok := err.(*noRetryErr); ok {
		err = e.err
	}
	return err
}
