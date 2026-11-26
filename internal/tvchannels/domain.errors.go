package tvchannels

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

type MissingResourceError string

func (e MissingResourceError) Error() string {
	return string(e)
}

type InfrastructureError string

func (e InfrastructureError) Error() string {
	return string(e)
}

const (
	ErrInvalidName               ValidationError      = "invalid name"
	ErrInvalidListPosition       ValidationError      = "invalid listPosition"
	ErrInvalidStatus             ValidationError      = "invalid status"
	ErrEmptySignals              ValidationError      = "empty signals"
	ErrInvalidSignalURL          ValidationError      = "invalid signals[].url"
	ErrInvalidSignalPriority     ValidationError      = "invalid signals[].priority"
	ErrInvalidRefreshRateSeconds ValidationError      = "invalid signals[].refreshRateSeconds"
	ErrInvalidBufferSize         ValidationError      = "invalid signals[].bufferSize"
	ErrInvalidID                 MissingResourceError = "invalid id"
	ErrInternal                  InfrastructureError  = "internal error"
	ErrInvalidStoredJson         InfrastructureError  = "invalid saved json"
	ErrDBInternal                InfrastructureError  = "db internal error"
)
