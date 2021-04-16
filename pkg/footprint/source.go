package footprint

type Source string

const (
	SourceNull          Source = ""
	SourceLogin         Source = "login"
	SourceSignUp        Source = "signup"
	SourceVerification  Source = "email_verification"
	SourcePasswordReset Source = "password_reset"
)
