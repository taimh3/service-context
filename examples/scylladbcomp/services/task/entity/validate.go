package entity

func ValidateTitle(title string) error {
	if title == "" {
		return ErrTitleCannotBeBlank
	}
	return nil
}

func ValidateStatus(status TaskType) error {
	if status == "" {
		return ErrStatusCannotBeBlank
	}

	if status != StatusDoing && status != StatusDone {
		return ErrStatusNotValid
	}

	return nil
}
