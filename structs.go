package main

// SearchRequest 定义搜索请求的结构
type SearchRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	OrderBy  string `json:"orderBy"`
	Order    string `json:"order"`
}

// CronjobItem 定义单个定时任务的结构
type CronjobItem struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Spec            string `json:"spec"`
	Script          string `json:"script"`
	Command         string `json:"command"`
	ContainerName   string `json:"containerName"`
	AppID           string `json:"appID"`
	Website         string `json:"website"`
	ExclusionRules  string `json:"exclusionRules"`
	DBType          string `json:"dbType"`
	DBName          string `json:"dbName"`
	URL             string `json:"url"`
	SourceDir       string `json:"sourceDir"`
	BackupAccounts  string `json:"backupAccounts"`
	DefaultDownload string `json:"defaultDownload"`
	RetainCopies    int    `json:"retainCopies"`
	LastRecordTime  string `json:"lastRecordTime"`
	Status          string `json:"status"`
	Secret          string `json:"secret"`
}

// SearchResponse 定义搜索响应的结构
type SearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Total int           `json:"total"`
		Items []CronjobItem `json:"items"`
	} `json:"data"`
}

// CronjobRecordRequest 定义获取定时任务记录的请求结构
type CronjobRecordRequest struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	CronjobID int `json:"cronjobID"`
	Days      int `json:"days"` // 获取从此时起往前多少天的数据
}

// CronjobRecordItem 定义单个定时任务记录的结构
type CronjobRecordItem struct {
	ID         int    `json:"id"`
	StartTime  string `json:"startTime"`
	Records    string `json:"records"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	TargetPath string `json:"targetPath"`
	Interval   int    `json:"interval"`
	File       string `json:"file"`
}

// CronjobRecordResponse 定义获取定时任务记录的响应结构
type CronjobRecordResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Total int                 `json:"total"`
		Items []CronjobRecordItem `json:"items"`
	} `json:"data"`
}
