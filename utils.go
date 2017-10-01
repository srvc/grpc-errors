package grpcerrors

func composeHandlers(funcs []ErrorHandlerFunc) ErrorHandlerFunc {
	return func(err error) error {
		if err != nil {
			for _, f := range funcs {
				err = f(err)
				if err == nil {
					break
				}
			}
		}
		return err
	}
}
