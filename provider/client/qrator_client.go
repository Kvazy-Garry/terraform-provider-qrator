package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// QRAPI — структура для работы с API Qrator
type QRAPI struct {
	Endpoint string
	Token    string
	ClientID int
	HttpCli  *http.Client
}

// NewQRClient — конструктор экземпляра клиента API
func NewQRClient(endpoint string, token string, clientID int) *QRAPI {
	return &QRAPI{
		Endpoint: endpoint,
		Token:    token,
		ClientID: clientID,
		HttpCli:  &http.Client{},
	}
}

// RPCRequest — структура запроса к API
type RPCRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	ID     int         `json:"id"`
}

// RPCResponse — структура ответа от API
type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     int         `json:"id"`
}

// SendRPCRequest — метод отправки запроса к API
func (cli *QRAPI) SendRPCRequest(method string, params interface{}, result interface{}) error {
	requestBody := RPCRequest{
		Method: method,
		Params: params,
		ID:     1,
	}

	jsonBytes, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%d", cli.Endpoint, cli.ClientID), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Qrator-Auth", cli.Token)
	resp, err := cli.HttpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response := RPCResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("%v", response.Error)
	}

	if result != nil {
		resultJson, err := json.Marshal(response.Result)
		if err != nil {
			return err
		}
		err = json.Unmarshal(resultJson, result)
		if err != nil {
			return err
		}
	}

	return nil
}
