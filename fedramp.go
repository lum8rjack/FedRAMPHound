package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

type AgencyProductRef struct {
	ID                string `json:"id,omitempty"`
	CSP               string `json:"csp,omitempty"`
	CSO               string `json:"cso,omitempty"`
	Status            string `json:"status,omitempty"`
	ImpactLevel       string `json:"impact_level,omitempty"`
	ImpactLevelNumber string `json:"impact_level_number,omitempty"`
}

type Agency struct {
	ID            string              `json:"id,omitempty"`
	Parent        string              `json:"parent,omitempty"`
	Sub           *string             `json:"sub,omitempty"`
	Logo          any                 `json:"logo,omitempty"`
	Authorization int                 `json:"authorization,omitempty"`
	Reuse         int                 `json:"reuse,omitempty"`
	Email         any                 `json:"email,omitempty"`
	Website       string              `json:"website,omitempty"`
	Auths         []AgencyProductRef  `json:"auths,omitempty"`
	Reuses        []AgencyProductRef  `json:"reuses,omitempty"`
	Procs         []AgencyProductRef  `json:"procs,omitempty"`
	FilterClasses string              `json:"filter_classes,omitempty"`
}

type AssessorClient struct {
	ID                string `json:"id,omitempty"`
	CSP               string `json:"csp,omitempty"`
	CSO               string `json:"cso,omitempty"`
	Status            string `json:"status,omitempty"`
	ImpactLevel       string `json:"impact_level,omitempty"`
	ImpactLevelNumber string `json:"impact_level_number,omitempty"`
}

type Assessor struct {
	ID                       string           `json:"id,omitempty"`
	Name                     string           `json:"name,omitempty"`
	Logo                     string           `json:"logo,omitempty"`
	ProductsAssessing        int              `json:"products_assessing,omitempty"`
	AccreditedSince          *time.Time       `json:"accredited_since,omitempty"`
	POC                      string           `json:"poc,omitempty"`
	Email                    string           `json:"email,omitempty"`
	Founded                  *time.Time       `json:"founded,omitempty"`
	Address                  string           `json:"address,omitempty"`
	Desc                     string           `json:"desc,omitempty"`
	Services                 string           `json:"services,omitempty"`
	CSPs                     []string         `json:"csps,omitempty"`
	Frameworks               []string         `json:"frameworks,omitempty"`
	Clients                  []AssessorClient `json:"clients,omitempty"`
	FilterClasses            string           `json:"filter_classes,omitempty"`
	HighestImpactLevel       string           `json:"highest_impact_level,omitempty"`
	HighestImpactLevelNumber int              `json:"highest_impact_level_number,omitempty"`
	ClientsInProcess         int              `json:"clients_in_process,omitempty"`
	Has20XAssessed           bool             `json:"has_20x_assessed,omitempty"`
}

type ProductEvent struct {
	Date        *time.Time `json:"date,omitempty"`
	Category    string     `json:"category,omitempty"`
	Description string     `json:"description,omitempty"`
}

type LeveragedSystem struct {
	ID                string `json:"id,omitempty"`
	CSP               string `json:"csp,omitempty"`
	CSO               string `json:"cso,omitempty"`
	Status            string `json:"status,omitempty"`
	ImpactLevel       string `json:"impact_level,omitempty"`
	ImpactLevelNumber string `json:"impact_level_number,omitempty"`
}

type Product struct {
	ID                             string            `json:"id,omitempty"`
	CSP                            string            `json:"csp,omitempty"`
	CSO                            string            `json:"cso,omitempty"`
	Logo                           string            `json:"logo,omitempty"`
	Status                         string            `json:"status,omitempty"`
	Phase                          string            `json:"phase,omitempty"`
	UnderCap                       bool              `json:"under_cap,omitempty"`
	CapDate                        any               `json:"cap_date,omitempty"`
	EventLog                       []ProductEvent    `json:"event_log,omitempty"`
	Authorization                  int               `json:"authorization,omitempty"`
	Reuse                          int               `json:"reuse,omitempty"`
	ReadyStatus                    string            `json:"ready_status,omitempty"`
	IPJabStatus                    string            `json:"ip_jab_status,omitempty"`
	IPProgStatus                   string            `json:"ip_prog_status,omitempty"`
	IPAgencyStatus                 string            `json:"ip_agency_status,omitempty"`
	IPPmoStatus                    string            `json:"ip_pmo_status,omitempty"`
	CertDate                       *time.Time        `json:"cert_date,omitempty"`
	StatusDate                     *time.Time        `json:"status_date,omitempty"`
	IPProgDate                     any               `json:"ip_prog_date,omitempty"`
	IPProgDate2                    any               `json:"ip_prog_date2,omitempty"`
	ReadyDate                      any               `json:"ready_date,omitempty"`
	IPPmoDate                      *time.Time        `json:"ip_pmo_date,omitempty"`
	IPAgencyDate                   *time.Time        `json:"ip_agency_date,omitempty"`
	IPJabDate                      any               `json:"ip_jab_date,omitempty"`
	CertPath                       string            `json:"cert_path,omitempty"`
	CertType                       string            `json:"cert_type,omitempty"`
	PartneringAgency               *string           `json:"partnering_agency,omitempty"`
	AnnualAssessment               *time.Time        `json:"annual_assessment,omitempty"`
	IndependentAssessor            string            `json:"independent_assessor,omitempty"`
	ServiceModel                   []string          `json:"service_model,omitempty"`
	DeploymentModel                string            `json:"deployment_model,omitempty"`
	ImpactLevel                    string            `json:"impact_level,omitempty"`
	ImpactLevelNumber              string            `json:"impact_level_number,omitempty"`
	LeveragedSystems               []LeveragedSystem `json:"leveraged_systems,omitempty"`
	AgencyAuthorizations           []string          `json:"agency_authorizations,omitempty"`
	AgencyReuse                    []string          `json:"agency_reuse,omitempty"`
	ServiceDesc                    string            `json:"service_desc,omitempty"`
	FedrampMsg                     string            `json:"fedramp_msg,omitempty"`
	SalesEmail                     string            `json:"sales_email,omitempty"`
	SecurityEmail                  string            `json:"security_email,omitempty"`
	Website                        string            `json:"website,omitempty"`
	UEI                            any               `json:"uei,omitempty"`
	SmallBusiness                  bool              `json:"small_business,omitempty"`
	BusinessCategories             []any             `json:"business_categories,omitempty"`
	ServiceLast90                  []string          `json:"service_last_90,omitempty"`
	AllOthers                      []string          `json:"all_others,omitempty"`
	FilterClasses                  string            `json:"filter_classes,omitempty"`
	ServiceAcronym                 string            `json:"service_acronym,omitempty"`
	TrustCenter                    any               `json:"trust_center,omitempty"`
	SecureConfigurationGuidance    any               `json:"secure_configuration_guidance,omitempty"`
	CertifiedServices              []any             `json:"certified_services,omitempty"`
	ThirdPartyInformationResources any               `json:"third_party_information_resources,omitempty"`
}

type FedRAMPData struct {
	Agencies  []Agency
	Assessors []Assessor
	Products  []Product
}

func loadFiles(agenciesPath, assessorsPath, productsPath string) (*FedRAMPData, error) {
	agenciesBytes, err := os.ReadFile(agenciesPath)
	if err != nil {
		return nil, err
	}
	assessorsBytes, err := os.ReadFile(assessorsPath)
	if err != nil {
		return nil, err
	}
	productsBytes, err := os.ReadFile(productsPath)
	if err != nil {
		return nil, err
	}

	data := &FedRAMPData{}
	if err := json.Unmarshal(agenciesBytes, &data.Agencies); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(assessorsBytes, &data.Assessors); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(productsBytes, &data.Products); err != nil {
		return nil, err
	}
	return data, nil
}

func agencyDisplayName(a Agency) string {
	if a.Sub != nil && strings.TrimSpace(*a.Sub) != "" {
		return *a.Sub
	}
	return a.Parent
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func asString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
