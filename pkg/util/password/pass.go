package utilpassword

import "golang.org/x/crypto/bcrypt"

func GeneratePasswordHash(password []byte) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func ComparePasswordHash(password, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}
