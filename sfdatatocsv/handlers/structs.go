package handlers

import "go.uber.org/zap"

type Handler struct {
	l       *zap.Logger
	cfg     *Config
	mcIdMap map[string]MeasureCalcRecords
}

type MeasureCalcsResponse struct {
	TotalSize          int                  `json:"totalSize"`
	Done               bool                 `json:"done"`
	MeasureCalcRecords []MeasureCalcRecords `json:"records"`
}

type MeasureCalcRecords struct {
	Attributes  Attributes `json:"attributes"`
	FieldToCalc string     `json:"CLI_CNR_Field_to_calculate__c"`
	Formula     string     `json:"CLR_CNI_Mesaure_Formula__c"`
	Sequence    float32    `json:"CLR_CNI_Sequence__c"`
	Id          string     `json:"Id"`
	Name        string     `json:"Name"`
	ProgramName string     `json:"Program_Name__c"`
	IsDeleted   bool       `json:"IsDeleted"`
}

type MCLIResponse struct {
	TotalSize   int           `json:"totalSize"`
	Done        bool          `json:"done"`
	MCLIRecords []MCLIRecords `json:"records"`
}

type MCLIRecords struct {
	Attributes  Attributes `json:"attributes"`
	Condition   string     `json:"Condition__c"`
	Formula     string     `json:"Measure_Formula__c"`
	Id          string     `json:"Id"`
	MeasureCalc string     `json:"Measure_Calculation__c"`
	IsDeleted   bool       `json:"IsDeleted"`
}

type Attributes struct {
	Type string `json:"type"`
}

type Config struct {
	JsonDirPath string
}
