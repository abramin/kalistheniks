package auth

// GenerateToken creates a signed JWT for the provided user.
func GenerateToken(userID, secret string) (string, error) {
	// TODO: implement JWT signing with proper claims and expiry.
	return "", nil
}

// ParseToken validates the JWT and extracts the user identifier.
func ParseToken(token, secret string) (string, error) {
	// TODO: implement JWT parsing and signature verification.
	return "", nil
}
