package handlers

import (
	"net/http"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Handler struct {
	clientID      string
	clientSecret  string
	username      string
	instanceURL   string
	authURL       string
	tokenURL      string
	sobjectsURL   string
	UserAgent     string
	accessToken   string
	jwtToken      string
	sfEnv         string
	pKeyPath      string
	queryURL      string
	uiapiURL      string
	uiapibatchURL string
	ingestURL     string

	l      *zap.Logger
	client *http.Client

	amqpCh    *amqp091.Channel
	amqpClose func() error
}

type FieldMetadata struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type MetadataResponse struct {
	Fields []FieldMetadata `json:"fields"`
}

type FieldAPILabelMapping map[string]string

type Payload struct {
	SObject             string                   `json:"sObject,omitempty"`
	FieldName           string                   `json:"fieldName,omitempty"`
	Query               string                   `json:"query,omitempty"`
	RecTypeID           string                   `json:"recTypeId,omitempty"`
	Records             []CustomRecords          `json:"records,omitempty"`
	PicklistMapToInsert PicklistMappedResp       `json:"picklist_map_insert,omitempty"`
	TargetSObject       string                   `json:"targetsObject,omitempty"`
	RecordsToInsert     []map[string]interface{} `json:"rec_to_insert,omitempty"`
}

type QueryResponse struct {
	TotalSize int             `json:"totalSize"`
	Done      bool            `json:"done"`
	Records   []CustomRecords `json:"records"`
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

type PicklistQueryResponse struct {
	PicklistValues []PicklistValue `json:"Values"`
}

type PicklistValue struct {
	PickValues string `json:"value"`
}

type CustomRecordsMap map[string][]CustomRecords

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

type BulkInsertResult struct {
	HasErrors bool        `json:"hasErrors,omitempty"`
	Results   interface{} `json:"results"`
}

type BulkInsert struct {
	AllOrNone  bool        `json:"allOrNone"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	Type    string   `json:"type"`
	Records []Record `json:"records"`
}

type Record struct {
	APIName string      `json:"apiName"`
	Fields  interface{} `json:"fields"`
}

type BulkCreateJobResult struct {
	ID              string  `json:"id"`
	Operation       string  `json:"operation"`
	Object          string  `json:"object"`
	CreatedByID     string  `json:"createdById,omitempty"`
	CreatedDate     string  `json:"createdDate,omitempty"`
	SystemModstamp  string  `json:"systemModstamp,omitempty"`
	State           string  `json:"state,omitempty"`
	ConcurrencyMode string  `json:"concurrencyMode,omitempty"`
	ContentType     string  `json:"contentType,omitempty"`
	APIVersion      float64 `json:"apiVersion,omitempty"`
	ContentURL      string  `json:"contentUrl,omitempty"`
	LineEnding      string  `json:"lineEnding,omitempty"`
	ColumnDelimiter string  `json:"columnDelimiter,omitempty"`
}
