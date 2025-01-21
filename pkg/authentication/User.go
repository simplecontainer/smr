package authentication

import "encoding/json"

func New(username string, domain string) *User {
	return &User{
		Username: username,
		Domain:   domain,
	}
}

func (user *User) FromString(str string) error {
	return json.Unmarshal([]byte(str), &user)
}

func (user *User) ToString() string {
	if user != nil {
		str, err := json.Marshal(user)

		if err != nil {
			return ""
		}

		return string(str)
	}

	return ""
}

func (user *User) ToBytes() []byte {
	if user != nil {
		str, err := json.Marshal(user)

		if err != nil {
			return nil
		}

		return str
	}

	return nil
}
