package errcode

// WrapGRPC converts the error to a gRPC status error while preserving
// successful results. It avoids repeating status conversion logic at call sites.
func WrapGRPC[T any](res T, err error) (T, error) {
	if err != nil {
		var zero T
		return zero, ToGRPCStatus(err)
	}
	return res, nil
}
