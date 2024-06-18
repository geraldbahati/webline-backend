package mpesa

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type MpesaPaymentRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

type MpesaPaymentResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

func InitiateMpesaPayment(token, businessShortCode, password, timestamp, amount, phoneNumber, callbackURL, accountReference, transactionDesc string) (*MpesaPaymentResponse, error) {
    client := &http.Client{}
    paymentRequest := MpesaPaymentRequest{
        BusinessShortCode: businessShortCode,
        Password:          password,
        Timestamp:         timestamp,
        TransactionType:   "CustomerPayBillOnline",
        Amount:            amount,
        PartyA:            phoneNumber,
        PartyB:            businessShortCode,
        PhoneNumber:       phoneNumber,
        CallBackURL:       callbackURL,
        AccountReference:  accountReference,
        TransactionDesc:   transactionDesc,
    }

    jsonRequest, err := json.Marshal(paymentRequest)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", mpesaBaseURL+"/mpesa/stkpush/v1/processrequest", bytes.NewBuffer(jsonRequest))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var paymentResponse MpesaPaymentResponse
    if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
        return nil, err
    }

    return &paymentResponse, nil
}
