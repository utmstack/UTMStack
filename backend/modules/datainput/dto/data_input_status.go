package dto

type CreateDataInputStatusRequest struct {
	ID        *string `json:"id"` // must be nil/absent — present → 400 idexists
	Source    string  `json:"source"`
	DataType  string  `json:"dataType"`
	Timestamp int64   `json:"timestamp"`
	Median    *int64  `json:"median"`
	Alias     *string `json:"alias"`
}

type UpdateDataInputStatusRequest struct {
	ID        string  `json:"id"`
	Source    string  `json:"source"`
	DataType  string  `json:"dataType"`
	Timestamp int64   `json:"timestamp"`
	Median    *int64  `json:"median"`
	Alias     *string `json:"alias"`
}

type DataInputStatusResponse struct {
	ID        string  `json:"id"`
	Source    string  `json:"source"`
	DataType  string  `json:"dataType"`
	Timestamp int64   `json:"timestamp"`
	Median    *int64  `json:"median"`
	Alias     *string `json:"alias"`
	IsDown    bool    `json:"isDown"`
}

type DataInputStatusCountFilters struct {
	IDEquals             *string `form:"id.equals"`
	IDContains           *string `form:"id.contains"`
	SourceEquals         *string `form:"source.equals"`
	SourceContains       *string `form:"source.contains"`
	DataTypeEquals       *string `form:"dataType.equals"`
	DataTypeContains     *string `form:"dataType.contains"`
	TimestampEquals      *int64  `form:"timestamp.equals"`
	TimestampGreaterThan *int64  `form:"timestamp.greaterThan"`
	TimestampLessThan    *int64  `form:"timestamp.lessThan"`
	MedianEquals         *int64  `form:"median.equals"`
	MedianGreaterThan    *int64  `form:"median.greaterThan"`
	MedianLessThan       *int64  `form:"median.lessThan"`
}

type DataInputStatusListFilters struct {
	Page int `form:"page"`
	Size int `form:"size"`
}
