package models

type SuccessAuthIntentResponse[T any] struct {
	Token           string
	UserInformation T
}
