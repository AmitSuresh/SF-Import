package rabbitmq

import "io"

type Config struct {
	AmqpUser    string
	AmqpPass    string
	AmqpHost    string
	AmqpPort    string
	JsonDirPath string
}

type PicklistQueueRequest struct {
	Method      string        `json:"Method"`
	Url         string        `json:"Url"`
	Body        io.Reader     `json:"Body"`
	AccessToken string        `json:"AccessToken"`
	CustomObj   CustomRecords `json:"record"`
	RecordType  string        `json:"recordType"`
}

type EquipmentRecord struct {
	ProgName    string `json:"Program_Name__c"`
	MeasureName string `json:"Measure_Description__c"`
	PicklistVal string `json:"Equipment_Type__c"`
}

type RecommendationRecord struct {
	ProgName    string `json:"Program_Name__c"`
	MeasureName string `json:"Measure_Description__c"`
	PicklistVal string `json:"Recommendation__c"`
}

type EquipmentRecords []EquipmentRecord

type RecommendationRecords []RecommendationRecord

type PicklistMappedResp struct {
	Eqs  EquipmentRecords      `json:"equipment_records,omitempty"`
	Recs RecommendationRecords `json:"recommendation_records,omitempty"`
}

type PicklistQueryResponse struct {
	PicklistValues []PicklistValue `json:"Values"`
}

type PicklistValue struct {
	PickValues string `json:"value"`
}

type CustomRecords struct {
	Id             string        `json:"Id"`
	MeasureNameNew string        `json:"Measure_Name_New__c,omitempty"`
	RecTypeName    string        `json:"Record_Type_Name__c,omitempty"`
	RecTypeId      string        `json:"Record_Type_Id__c,omitempty"`
	ProgRec        ProgramRecord `json:"Program__r,omitempty"`
}

type ProgramRecord struct {
	Name string `json:"Name,omitempty"`
}
