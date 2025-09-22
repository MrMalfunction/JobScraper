package api_models

type StdResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
