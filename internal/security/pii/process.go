package pii

type ProcessResult struct {
	Normalized string
	Hash       string
	Encrypted  string
}

func Process(value string, key []byte) (ProcessResult, error) {
	normalized := Normalize(value)
	encrypted, err := Encrypt(normalized, key)
	if err != nil {
		return ProcessResult{}, err
	}

	return ProcessResult{
		Normalized: normalized,
		Hash:       Hash(normalized),
		Encrypted:  encrypted,
	}, nil
}
