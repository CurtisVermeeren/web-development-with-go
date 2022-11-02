package views

import "log"

// Data is a top level structure that is used to send data to views
type Data struct {
	Alert *Alert
	Yield interface{}
}

// Alert is used to render Bootstrap alert messages in templated
type Alert struct {
	Level   string
	Message string
}

// Represent Bootstrap Alerts in Go
const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists."
)

// PublicError is used to generate error messages safe for the public
type PublicError interface {
	error
	Public() string
}

// SetAlert adds an alert to a Data object
func (d *Data) SetAlert(err error) {
	var msg string

	// Check if the error is for Public or if it should be generic
	if pErr, ok := err.(PublicError); ok {
		msg = pErr.Public()
	} else {
		log.Println(err)
		msg = AlertMsgGeneric
	}
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

// AlertError sets a custom message on the Data Alert object
func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}
