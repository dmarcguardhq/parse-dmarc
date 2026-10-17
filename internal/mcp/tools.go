package mcp

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/dmarcguardhq/parse-dmarc/internal/parser"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool input types

// EmptyInput is used for tools that don't require any input.
type EmptyInput struct{}

// PaginationInput is used for tools that support pagination.
type PaginationInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"maximum number of results to return (default: 50)"`
	Offset int `json:"offset,omitempty" jsonschema:"number of results to skip for pagination (default: 0)"`
}

// ReportIDInput is used for tools that require a report ID.
type ReportIDInput struct {
	ID int64 `json:"id" jsonschema:"the database ID of the report to retrieve"`
}

// LimitInput is used for tools that only need a limit parameter.
type LimitInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"maximum number of results to return (default: 10)"`
}

// ParseReportInput is used for parsing raw DMARC reports.
type ParseReportInput struct {
	ReportData string `json:"report_data" jsonschema:"base64 encoded DMARC report data (gzip/zip/XML)"`
}

// Tool output types

// StatisticsOutput wraps the statistics response.
type StatisticsOutput struct {
	Statistics *storage.Statistics `json:"statistics"`
}

// ReportsOutput wraps the reports list response.
type ReportsOutput struct {
	Reports []storage.ReportSummary `json:"reports"`
	Count   int                     `json:"count"`
}

// ReportOutput wraps a single report response.
type ReportOutput struct {
	Report *parser.Feedback `json:"report"`
}

// TopSourceIPsOutput wraps the top source IPs response.
type TopSourceIPsOutput struct {
	SourceIPs []storage.TopSourceIP `json:"source_ips"`
	Count     int                   `json:"count"`
}

// DomainStatsOutput wraps domain statistics response.
type DomainStatsOutput struct {
	Domains []storage.DomainStats `json:"domains"`
	Count   int                   `json:"count"`
}

// OrgStatsOutput wraps organization statistics response.
type OrgStatsOutput struct {
	Organizations []storage.OrgStats `json:"organizations"`
	Count         int                `json:"count"`
}

// AuthResultStatsOutput wraps authentication result statistics response.
type AuthResultStatsOutput struct {
	Results []storage.AuthResultStats `json:"results"`
	Count   int                       `json:"count"`
}

// ParsedReportOutput wraps a parsed DMARC report response.
type ParsedReportOutput struct {
	Report         *parser.Feedback `json:"report"`
	TotalMessages  int              `json:"total_messages"`
	CompliantCount int              `json:"compliant_count"`
	ComplianceRate float64          `json:"compliance_rate"`
	DateBegin      string           `json:"date_begin"`
	DateEnd        string           `json:"date_end"`
}

// Tool handlers

func (s *Server) getStatistics(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, StatisticsOutput, error) {
	stats, err := s.store.GetStatistics()
	if err != nil {
		return nil, StatisticsOutput{}, fmt.Errorf("failed to get statistics: %w", err)
	}

	return nil, StatisticsOutput{Statistics: stats}, nil
}

func (s *Server) getReports(ctx context.Context, req *mcp.CallToolRequest, input PaginationInput) (*mcp.CallToolResult, ReportsOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	reports, err := s.store.GetReports(limit, offset)
	if err != nil {
		return nil, ReportsOutput{}, fmt.Errorf("failed to get reports: %w", err)
	}

	if reports == nil {
		reports = []storage.ReportSummary{}
	}

	return nil, ReportsOutput{
		Reports: reports,
		Count:   len(reports),
	}, nil
}

func (s *Server) getReportByID(ctx context.Context, req *mcp.CallToolRequest, input ReportIDInput) (*mcp.CallToolResult, ReportOutput, error) {
	if input.ID <= 0 {
		return nil, ReportOutput{}, fmt.Errorf("invalid report ID: must be a positive integer")
	}

	report, err := s.store.GetReportByID(input.ID)
	if err != nil {
		return nil, ReportOutput{}, fmt.Errorf("failed to get report: %w", err)
	}

	// Normalize slice fields to ensure valid JSON schema compliance
	// (nil slices serialize as null, but schema expects arrays)
	report.NormalizeForJSON()

	return nil, ReportOutput{Report: report}, nil
}

func (s *Server) getTopSourceIPs(ctx context.Context, req *mcp.CallToolRequest, input LimitInput) (*mcp.CallToolResult, TopSourceIPsOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	ips, err := s.store.GetTopSourceIPs(limit)
	if err != nil {
		return nil, TopSourceIPsOutput{}, fmt.Errorf("failed to get top source IPs: %w", err)
	}

	if ips == nil {
		ips = []storage.TopSourceIP{}
	}

	return nil, TopSourceIPsOutput{
		SourceIPs: ips,
		Count:     len(ips),
	}, nil
}

func (s *Server) getDomainStats(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, DomainStatsOutput, error) {
	stats, err := s.store.GetDomainStats()
	if err != nil {
		return nil, DomainStatsOutput{}, fmt.Errorf("failed to get domain stats: %w", err)
	}

	if stats == nil {
		stats = []storage.DomainStats{}
	}

	return nil, DomainStatsOutput{
		Domains: stats,
		Count:   len(stats),
	}, nil
}

func (s *Server) getOrgStats(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, OrgStatsOutput, error) {
	stats, err := s.store.GetOrgStats()
	if err != nil {
		return nil, OrgStatsOutput{}, fmt.Errorf("failed to get organization stats: %w", err)
	}

	if stats == nil {
		stats = []storage.OrgStats{}
	}

	return nil, OrgStatsOutput{
		Organizations: stats,
		Count:         len(stats),
	}, nil
}

func (s *Server) getSPFStats(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, AuthResultStatsOutput, error) {
	stats, err := s.store.GetSPFStats()
	if err != nil {
		return nil, AuthResultStatsOutput{}, fmt.Errorf("failed to get SPF stats: %w", err)
	}

	if stats == nil {
		stats = []storage.AuthResultStats{}
	}

	return nil, AuthResultStatsOutput{
		Results: stats,
		Count:   len(stats),
	}, nil
}

func (s *Server) getDKIMStats(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, AuthResultStatsOutput, error) {
	stats, err := s.store.GetDKIMStats()
	if err != nil {
		return nil, AuthResultStatsOutput{}, fmt.Errorf("failed to get DKIM stats: %w", err)
	}

	if stats == nil {
		stats = []storage.AuthResultStats{}
	}

	return nil, AuthResultStatsOutput{
		Results: stats,
		Count:   len(stats),
	}, nil
}

func (s *Server) parseDMARCReport(ctx context.Context, req *mcp.CallToolRequest, input ParseReportInput) (*mcp.CallToolResult, ParsedReportOutput, error) {
	if input.ReportData == "" {
		return nil, ParsedReportOutput{}, fmt.Errorf("report_data is required")
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(input.ReportData)
	if err != nil {
		return nil, ParsedReportOutput{}, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	const maxReportSize = 10 * 1024 * 1024 // 10MB
	if len(data) > maxReportSize {
		return nil, ParsedReportOutput{}, fmt.Errorf("report data exceeds maximum size of %d bytes", maxReportSize)
	}
	// Parse the report
	report, err := parser.ParseReport(data)
	if err != nil {
		return nil, ParsedReportOutput{}, fmt.Errorf("failed to parse DMARC report: %w", err)
	}

	// Normalize slice fields to ensure valid JSON schema compliance
	// (nil slices serialize as null, but schema expects arrays)
	report.NormalizeForJSON()

	// Calculate statistics
	totalMessages := report.GetTotalMessages()
	compliantCount := report.GetDMARCCompliantCount()
	var complianceRate float64
	if totalMessages > 0 {
		complianceRate = float64(compliantCount) / float64(totalMessages) * 100
	}

	// Get date range
	begin, end := report.GetDateRange()

	return nil, ParsedReportOutput{
		Report:         report,
		TotalMessages:  totalMessages,
		CompliantCount: compliantCount,
		ComplianceRate: complianceRate,
		DateBegin:      begin.Format("2006-01-02 15:04:05 UTC"),
		DateEnd:        end.Format("2006-01-02 15:04:05 UTC"),
	}, nil
}
