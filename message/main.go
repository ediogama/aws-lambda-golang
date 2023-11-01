package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"net/http"

	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"

	owm "github.com/briandowns/openweathermap"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var bodyMessage []string

	body := strings.Split(request.Body, "&")

	for e := range body {
		if strings.Contains(body[e], "Body") {
			bodyMessage = strings.Split(body[e], "=")
		}
	}

	w, err := owm.NewCurrent("C", "pt", os.Getenv("OWM_API_KEY"))
	if err != nil {
		log.Fatalln(err)
	}

	city := bodyMessage[1]
	w.CurrentByName(city)

	var bufMSG bytes.Buffer
	// Set environment variables
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")

	// Add proxy settings to a http Transport object
	transport := &http.Transport{
		// https://pkg.go.dev/net/http#ProxyFromEnvironment
		Proxy: http.ProxyFromEnvironment,
	}

	// Add the Transport to an http Client
	httpClient := &http.Client{
		Transport: transport,
	}

	// Create your custom Twilio client using the http client and your credentials
	twilioHttpClient := client.Client{
		Credentials: client.NewCredentials(accountSid, authToken),
		HTTPClient:  httpClient,
	}
	twilioHttpClient.SetAccountSid(accountSid)
	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{Client: &twilioHttpClient})

	params := &twilioApi.CreateMessageParams{}
	params.SetFrom("whatsapp:+14155238886")
	params.SetBody(fmt.Sprintf("A temperatura agora em %v é %v graus U+1F916\n", city, w.Main.Temp))
	params.SetTo("whatsapp:+5511948149113")

	resp1, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		return Response{StatusCode: 404}, err
	} else {
		response, err := json.Marshal(*resp1)
		if err != nil {
			return Response{StatusCode: 404}, err
		}
		json.HTMLEscape(&bufMSG, response)
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            bufMSG.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "message-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
