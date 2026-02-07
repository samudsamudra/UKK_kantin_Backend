package app

// LoginErrorResponse error khusus login (aman & user friendly)
func LoginErrorResponse() map[string]string {
	return map[string]string{
		"error": "email atau password salah",
	}
}
