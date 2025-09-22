package api_models

import "encoding/json"

// type ListChatsRequest struct {
// 	Page     int `query:"page" validate:"omitempty,min=1" default:"1"`
// 	PageSize int `query:"page_size" validate:"omitempty,min=1,max=100" default:"1"`
// }

type AddWorkdayCompanyScrapeList struct {
	Name       string          `json:"name"`
	BaseUrl    string          `json:"base_url"`
	ApiReqBody json.RawMessage `json:"req_body"`
}
