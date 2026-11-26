package tvchannels

type TvSignal struct {
	URL                string
	Priority           int
	NeedBrowser        bool
	NeedRefresh        bool
	RefreshRateSeconds int
	BufferEnabled      bool
	BufferSize         int
}

func (sig *TvSignal) Empty() bool {
	return sig.URL == "" && sig.Priority == 0 && !sig.NeedBrowser
}

func (sig *TvSignal) Validate() error {
	if sig.Empty() {
		return nil
	}

	if sig.URL == "" {
		return ErrInvalidSignalURL
	}

	if sig.Priority < 1 {
		return ErrInvalidSignalPriority
	}

	if sig.NeedRefresh && sig.RefreshRateSeconds < 5 {
		return ErrInvalidRefreshRateSeconds
	}

	if sig.BufferEnabled && sig.BufferSize <= 0 {
		return ErrInvalidBufferSize
	}

	return nil
}

type TvSignalSet [5]TvSignal

func (set TvSignalSet) Empty() bool {
	isEmpty := true
	for _, tvsignal := range set {
		isEmpty = isEmpty && tvsignal.Empty()
	}

	return isEmpty
}

func (set TvSignalSet) Validate() error {
	for _, signal := range set {
		if signal.Empty() {
			continue
		}

		if err := signal.Validate(); err != nil {
			return err
		}
	}

	return nil
}
