package dashboard

import (
	"encoding/json"
	"fmt"
)

type Dashboard struct {
	Annotations          Annotations   `json:"annotations"`
	Description          string        `json:"description"`
	Editable             bool          `json:"editable"`
	FiscalYearStartMonth int           `json:"fiscalYearStartMonth"`
	GraphTooltip         int           `json:"graphTooltip"`
	ID                   int           `json:"id"`
	Links                []interface{} `json:"links"`
	LiveNow              bool          `json:"liveNow"`
	Panels               []Panel       `json:"panels"`
	Refresh              string        `json:"refresh"`
	SchemaVersion        int           `json:"schemaVersion"`
	Style                string        `json:"style"`
	Tags                 []string      `json:"tags"`
	Templating           Templating    `json:"templating"`
	Time                 TimeRange     `json:"time"`
	Timepicker           Timepicker    `json:"timepicker"`
	Timezone             string        `json:"timezone"`
	Title                string        `json:"title"`
	UID                  string        `json:"uid"`
	Version              int           `json:"version"`
	WeekStart            string        `json:"weekStart"`
}

type Annotations struct {
	List []AnnotationList `json:"list"`
}

type AnnotationList struct {
	BuiltIn    int        `json:"builtIn"`
	Datasource Datasource `json:"datasource"`
	Enable     bool       `json:"enable"`
	Hide       bool       `json:"hide"`
	IconColor  string     `json:"iconColor"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
}

type Datasource struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

type Panel struct {
	Datasource  Datasource  `json:"datasource"`
	Description string      `json:"description"`
	FieldConfig FieldConfig `json:"fieldConfig"`
	GridPos     GridPos     `json:"gridPos"`
	ID          int         `json:"id"`
	Options     Options     `json:"options"`
	Targets     []Target    `json:"targets"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
}

type FieldConfig struct {
	Defaults  Defaults      `json:"defaults"`
	Overrides []interface{} `json:"overrides"`
}

type Defaults struct {
	Color      Color         `json:"color"`
	Custom     Custom        `json:"custom"`
	Decimals   int           `json:"decimals"`
	Mappings   []interface{} `json:"mappings"`
	Thresholds Thresholds    `json:"thresholds"`
	Unit       string        `json:"unit"`
}

type Color struct {
	Mode string `json:"mode"`
}

type Custom struct {
	AxisCenteredZero  bool              `json:"axisCenteredZero"`
	AxisColorMode     string            `json:"axisColorMode"`
	AxisLabel         string            `json:"axisLabel"`
	AxisPlacement     string            `json:"axisPlacement"`
	AxisSoftMax       int               `json:"axisSoftMax"`
	AxisSoftMin       int               `json:"axisSoftMin"`
	BarAlignment      int               `json:"barAlignment"`
	DrawStyle         string            `json:"drawStyle"`
	FillOpacity       int               `json:"fillOpacity"`
	GradientMode      string            `json:"gradientMode"`
	HideFrom          HideFrom          `json:"hideFrom"`
	InsertNulls       bool              `json:"insertNulls"`
	LineInterpolation string            `json:"lineInterpolation"`
	LineWidth         int               `json:"lineWidth"`
	PointSize         int               `json:"pointSize"`
	ScaleDistribution ScaleDistribution `json:"scaleDistribution"`
	ShowPoints        string            `json:"showPoints"`
	SpanNulls         bool              `json:"spanNulls"`
	Stacking          Stacking          `json:"stacking"`
	ThresholdsStyle   ThresholdsStyle   `json:"thresholdsStyle"`
}

type HideFrom struct {
	Legend  bool `json:"legend"`
	Tooltip bool `json:"tooltip"`
	Viz     bool `json:"viz"`
}

type ScaleDistribution struct {
	Type string `json:"type"`
}

type Stacking struct {
	Group string `json:"group"`
	Mode  string `json:"mode"`
}

type ThresholdsStyle struct {
	Mode string `json:"mode"`
}

type Thresholds struct {
	Mode  string `json:"mode"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Color string      `json:"color"`
	Value interface{} `json:"value"`
}

type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

type Options struct {
	Legend  Legend  `json:"legend"`
	Tooltip Tooltip `json:"tooltip"`
}

type Legend struct {
	Calcs       []string `json:"calcs"`
	DisplayMode string   `json:"displayMode"`
	Placement   string   `json:"placement"`
	ShowLegend  bool     `json:"showLegend"`
}

type Tooltip struct {
	Mode string `json:"mode"`
	Sort string `json:"sort"`
}

type Target struct {
	Datasource   Datasource `json:"datasource"`
	EditorMode   string     `json:"editorMode"`
	Expr         string     `json:"expr"`
	Instant      bool       `json:"instant"`
	LegendFormat string     `json:"legendFormat"`
	Range        bool       `json:"range"`
	RefId        string     `json:"refId"`
}

type Templating struct {
	List []TemplatingDetail `json:"list"`
}

type TemplatingDetail struct {
	Current     CurrentDetail `json:"current"`
	Hide        int           `json:"hide"`
	IncludeAll  bool          `json:"includeAll"`
	Label       string        `json:"label"`
	Multi       bool          `json:"multi"`
	Name        string        `json:"name"`
	Options     []interface{} `json:"options"`
	Query       QueryValue    `json:"query"`
	QueryValue  string        `json:"queryValue"`
	Refresh     int           `json:"refresh"`
	Regex       string        `json:"regex"`
	SkipUrlSync bool          `json:"skipUrlSync"`
	Type        string        `json:"type"`
}

type CurrentDetail struct {
	Selected bool   `json:"selected"`
	Text     string `json:"text"`
	Value    string `json:"value"`
}

type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Timepicker struct{}

type QueryObject struct {
	Query string `json:"query"`
	Refid string `json:"refid"`
}

type QueryValue struct {
	StringValue string
	ObjectValue *QueryObject
	IsString    bool
}

func (qv *QueryValue) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		qv.StringValue = s
		qv.IsString = true
		return nil
	}

	var obj QueryObject
	if err := json.Unmarshal(data, &obj); err == nil {
		qv.ObjectValue = &obj
		qv.IsString = false
		return nil
	}

	return fmt.Errorf("failed to unmarshal QueryValue")
}

// Implement the json.Marshaler interface for QueryValue
func (qv *QueryValue) MarshalJSON() ([]byte, error) {
	if qv.IsString {
		return json.Marshal(qv.StringValue)
	}
	return json.Marshal(qv.ObjectValue)
}
