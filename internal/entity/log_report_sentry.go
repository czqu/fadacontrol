package entity

type LogReportSentry struct {
	UserId      string `gorm:"not null;uniqueIndex:idx_log_report_sentry"`
	Enable      bool   `gorm:"not null;default:true"`
	ReportLevel string `gorm:"not null;default:fatal"`
}
