package main

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

// ZTEMessage represents the raw message structure from the router
type ZTEMessage struct {
	ID                   string `json:"id"`
	Number               string `json:"number"`
	Content              string `json:"content"`
	Tag                  string `json:"tag"`
	Date                 string `json:"date"`
	ReceivedAllConcatSMS string `json:"received_all_concat_sms"`
	ConcatSMSTotal       string `json:"concat_sms_total"`
	ConcatSMSReceived    string `json:"concat_sms_received"`
	SMSClass             string `json:"sms_class"`
}

// ZTESMS represents a parsed SMS message
type ZTESMS struct {
	ID                   int       `json:"id"`
	Number               string    `json:"number"`
	Content              string    `json:"content"`
	Tag                  string    `json:"tag"`
	Date                 time.Time `json:"date"`
	ReceivedAllConcatSMS bool      `json:"received_all_concat_sms"`
	ConcatSMSTotal       int       `json:"concat_sms_total"`
	ConcatSMSReceived    int       `json:"concat_sms_received"`
	SMSClass             int       `json:"sms_class"`

	Read bool `json:"read"` // True = Read, False = Unread
}

func NewZTESMS(msg ZTEMessage) (ZTESMS, error) {
	sms := ZTESMS{
		Number: msg.Number,
		Tag:    msg.Tag,
	}

	// Parse ID
	id, err := strconv.Atoi(msg.ID)
	if err != nil {
		return sms, fmt.Errorf("invalid ID: %w", err)
	}
	sms.ID = id

	// Decode content from hex and UTF-16BE to UTF-8
	contentBytes, err := hex.DecodeString(msg.Content)
	if err != nil {
		return sms, fmt.Errorf("failed to decode content hex: %w", err)
	}

	// Convert from UTF-16BE to UTF-8
	var contentUTF16 []uint16
	for i := 0; i < len(contentBytes); i += 2 {
		if i+1 >= len(contentBytes) {
			break
		}
		contentUTF16 = append(contentUTF16, uint16(contentBytes[i])<<8|uint16(contentBytes[i+1]))
	}
	sms.Content = string(utf16.Decode(contentUTF16))

	// Parse date
	dateParts := strings.Split(msg.Date, ",")
	if len(dateParts) >= 6 {
		year, _ := strconv.Atoi(dateParts[0])
		month, _ := strconv.Atoi(dateParts[1])
		day, _ := strconv.Atoi(dateParts[2])
		hour, _ := strconv.Atoi(dateParts[3])
		min, _ := strconv.Atoi(dateParts[4])
		sec, _ := strconv.Atoi(dateParts[5])
		sms.Date = time.Date(2000+year, time.Month(month), day, hour, min, sec, 0, time.Local)
	} else {
		sms.Date = time.Now()
	}

	// Determine if the message is read or unread
	sms.Read = msg.Tag != "1" // Tag "1" = Unread, otherwise Read

	// Parse booleans and integers
	sms.ReceivedAllConcatSMS = msg.ReceivedAllConcatSMS == "1"
	sms.ConcatSMSTotal, _ = strconv.Atoi(msg.ConcatSMSTotal)
	sms.ConcatSMSReceived, _ = strconv.Atoi(msg.ConcatSMSReceived)
	sms.SMSClass, _ = strconv.Atoi(msg.SMSClass)

	return sms, nil
}
