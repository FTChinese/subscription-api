package conv

func LiveMode(b bool) string {
	if b {
		return "live"
	}

	return "sandbox"
}
