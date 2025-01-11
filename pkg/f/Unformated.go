package f

func NewUnformated(key string, category string) Unformated {
	return Unformated{
		Key:      key,
		Category: category,
		Type:     TYPE_UNFORMATED,
	}
}

func (format Unformated) GetCategory() string {
	return format.Category
}

func (format Unformated) GetType() string {
	return format.Type
}

func (format Unformated) IsValid() bool {
	return format.Key != ""
}

func (format Unformated) Full() bool {
	return format.Key != ""
}

func (format Unformated) ToString() string {
	return format.Key
}

func (format Unformated) ToBytes() []byte {
	return []byte(format.Key)
}
