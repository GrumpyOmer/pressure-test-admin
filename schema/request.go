package schema

type PublicRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ConcurrencyNum int64  `json:"concurrency_num"`
		Success        int64  `json:"success"`
		Fail           int64  `json:"fail"`
		Qps            string `json:"qps"`
		MaxTime        string `json:"max_time"`
		MinTime        string `json:"min_time"`
		AvgTime        string `json:"avg_time"`
		CodeInfo       string `json:"code_info"`
	} `json:"data"`
}

type PressureByUrlReq struct {
	Url                 string `json:"url"`
	ConcurrencyQuantity uint64 `json:"concurrency_quantity"`
}

type PressureByCurlReq struct {
	ConcurrencyQuantity uint64 `json:"concurrency_quantity"`
}

type PressureByGolangReq struct {
	ConcurrencyQuantity uint64 `json:"concurrency_quantity"`
	Port                int    `json:"port"`
}
