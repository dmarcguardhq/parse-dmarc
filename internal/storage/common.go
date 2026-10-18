package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/goccy/go-json"

	"github.com/dmarcguardhq/parse-dmarc/internal/parser"
)

type Storage struct {
	db *sql.DB
}

type ReportSummary struct {
	ID                int64  `json:"id"`
	ReportID          string `json:"report_id"`
	OrgName           string `json:"org_name"`
	Domain            string `json:"domain"`
	DateBegin         int64  `json:"date_begin"`
	DateEnd           int64  `json:"date_end"`
	TotalMessages     int    `json:"total_messages"`
	CompliantMessages int    `json:"compliant_messages"`
	// ComplianceRate is null for an RFC 7489 null report (no messages):
	// 0 of 0 is not a 0% pass rate.
	ComplianceRate *float64 `json:"compliance_rate"`
	PolicyP        string   `json:"policy_p"`
	// Status is the backend's verdict for display: empty, pass, warn or fail.
	Status string `json:"status"`
}

// Compliance-rate thresholds (percent) behind Status.
const (
	passRate = 100
	warnRate = 80
)

// classify derives ComplianceRate and Status from the message counts.
func (r *ReportSummary) classify() {
	if r.TotalMessages == 0 {
		r.Status = "empty"
		return
	}
	rate := float64(r.CompliantMessages) / float64(r.TotalMessages) * 100
	r.ComplianceRate = &rate
	switch {
	case rate >= passRate:
		r.Status = "pass"
	case rate >= warnRate:
		r.Status = "warn"
	default:
		r.Status = "fail"
	}
}

type Statistics struct {
	TotalReports      int     `json:"total_reports"`
	TotalMessages     int     `json:"total_messages"`
	CompliantMessages int     `json:"compliant_messages"`
	ComplianceRate    float64 `json:"compliance_rate"`
	// EnforcedMessages counts non-compliant mail the receiver already blocked
	// (disposition reject or quarantine): the published policy working, not
	// an authentication gap.
	EnforcedMessages int `json:"enforced_messages"`
	// DeliveredComplianceRate is the pass rate over mail that was actually
	// delivered (total minus enforced). Null when nothing was delivered.
	DeliveredComplianceRate *float64 `json:"delivered_compliance_rate"`
	// Health is the backend's verdict for the dashboard: nodata, secure,
	// warning or critical.
	Health          string `json:"health"`
	UniqueSourceIPs int    `json:"unique_source_ips"`
	UniqueDomains   int    `json:"unique_domains"`
	HasData         bool   `json:"has_data"`
}

// Delivered-compliance thresholds (percent) behind Health.
const (
	secureRate  = 95
	warningRate = 80
)

// classify derives DeliveredComplianceRate and Health from the counts.
func (st *Statistics) classify() {
	if !st.HasData {
		st.Health = "nodata"
		return
	}
	delivered := st.TotalMessages - st.EnforcedMessages
	if delivered <= 0 {
		// Everything seen was blocked spoofing: nothing unauthenticated got through.
		st.Health = "secure"
		return
	}
	rate := float64(st.CompliantMessages) / float64(delivered) * 100
	st.DeliveredComplianceRate = &rate
	switch {
	case rate >= secureRate:
		st.Health = "secure"
	case rate >= warningRate:
		st.Health = "warning"
	default:
		st.Health = "critical"
	}
}

type TopSourceIP struct {
	SourceIP string `json:"source_ip"`
	Count    int    `json:"count"`
	Pass     int    `json:"pass"`
	Fail     int    `json:"fail"`
}

func (s *Storage) SaveReport(feedback *parser.Feedback) error {
	rawReport, err := json.Marshal(feedback)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(`
		INSERT OR IGNORE INTO reports (
			report_id, org_name, email, domain,
			date_begin, date_end, created_at,
			policy_p, policy_sp, policy_pct,
			total_messages, compliant_messages,
			raw_report
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		feedback.ReportMetadata.ReportID,
		feedback.ReportMetadata.OrgName,
		feedback.ReportMetadata.Email,
		feedback.PolicyPublished.Domain,
		feedback.ReportMetadata.DateRange.Begin,
		feedback.ReportMetadata.DateRange.End,
		time.Now().Unix(),
		feedback.PolicyPublished.P,
		feedback.PolicyPublished.SP,
		feedback.PolicyPublished.PCT,
		feedback.GetTotalMessages(),
		feedback.GetDMARCCompliantCount(),
		rawReport,
	)

	if err != nil {
		return fmt.Errorf("failed to insert report: %w", err)
	}

	reportID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert ID: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil
	}

	for _, record := range feedback.Records {
		dkimDomains, _ := json.Marshal(record.AuthResults.DKIM)
		spfDomains, _ := json.Marshal(record.AuthResults.SPF)

		_, err := tx.Exec(`
			INSERT INTO records (
				report_id, source_ip, count,
				disposition, dkim_result, spf_result,
				header_from, envelope_from,
				dkim_domains, spf_domains
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			reportID,
			record.Row.SourceIP,
			record.Row.Count,
			record.Row.PolicyEvaluated.Disposition,
			record.Row.PolicyEvaluated.DKIM,
			record.Row.PolicyEvaluated.SPF,
			record.Identifiers.HeaderFrom,
			record.Identifiers.EnvelopeFrom,
			dkimDomains,
			spfDomains,
		)

		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// MaxReportLimit caps the page size GetReports callers may ask for, so one
// request cannot buffer the whole table into memory.
const MaxReportLimit = 100

func (s *Storage) GetReports(limit, offset int) ([]ReportSummary, error) {
	rows, err := s.db.Query(`
		SELECT id, report_id, org_name, domain,
		       date_begin, date_end,
		       total_messages, compliant_messages,
		       policy_p
		FROM reports
		ORDER BY date_begin DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reports []ReportSummary
	for rows.Next() {
		var r ReportSummary
		err := rows.Scan(
			&r.ID, &r.ReportID, &r.OrgName, &r.Domain,
			&r.DateBegin, &r.DateEnd,
			&r.TotalMessages, &r.CompliantMessages,
			&r.PolicyP,
		)
		if err != nil {
			return nil, fmt.Errorf("scan report row: %w", err)
		}

		r.classify()
		reports = append(reports, r)
	}

	return reports, nil
}

func (s *Storage) GetReportByID(id int64) (*parser.Feedback, error) {
	var rawReport string
	err := s.db.QueryRow("SELECT raw_report FROM reports WHERE id = ?", id).Scan(&rawReport)
	if err != nil {
		return nil, fmt.Errorf("query report %d: %w", id, err)
	}

	var feedback parser.Feedback
	if err := json.Unmarshal([]byte(rawReport), &feedback); err != nil {
		return nil, fmt.Errorf("unmarshal report %d: %w", id, err)
	}

	return &feedback, nil
}

func (s *Storage) GetStatistics() (*Statistics, error) {
	var stats Statistics

	err := s.db.QueryRow(`
		SELECT
			COUNT(*) as total_reports,
			COALESCE(SUM(total_messages), 0) as total_messages,
			COALESCE(SUM(compliant_messages), 0) as compliant_messages
		FROM reports
	`).Scan(&stats.TotalReports, &stats.TotalMessages, &stats.CompliantMessages)

	if err != nil {
		return nil, fmt.Errorf("query report statistics: %w", err)
	}

	stats.HasData = stats.TotalReports > 0

	if stats.TotalMessages > 0 {
		stats.ComplianceRate = float64(stats.CompliantMessages) / float64(stats.TotalMessages) * 100
	}

	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(count), 0)
		FROM records
		WHERE disposition IN ('reject', 'quarantine')
		  AND dkim_result != 'pass'
		  AND spf_result != 'pass'
	`).Scan(&stats.EnforcedMessages)
	if err != nil {
		return nil, fmt.Errorf("query enforced messages: %w", err)
	}

	err = s.db.QueryRow("SELECT COUNT(DISTINCT source_ip) FROM records").Scan(&stats.UniqueSourceIPs)
	if err != nil {
		return nil, fmt.Errorf("query unique source IPs: %w", err)
	}

	err = s.db.QueryRow("SELECT COUNT(DISTINCT domain) FROM reports").Scan(&stats.UniqueDomains)
	if err != nil {
		return nil, fmt.Errorf("query unique domains: %w", err)
	}

	stats.classify()
	return &stats, nil
}

func (s *Storage) GetTopSourceIPs(limit int) ([]TopSourceIP, error) {
	rows, err := s.db.Query(`
		SELECT
			source_ip,
			SUM(count) as total_count,
			SUM(CASE WHEN (dkim_result = 'pass' OR spf_result = 'pass') THEN count ELSE 0 END) as pass_count,
			SUM(CASE WHEN (dkim_result != 'pass' AND spf_result != 'pass') THEN count ELSE 0 END) as fail_count
		FROM records
		GROUP BY source_ip
		ORDER BY total_count DESC
		LIMIT ?
	`, limit)

	if err != nil {
		return nil, fmt.Errorf("query top source IPs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []TopSourceIP
	for rows.Next() {
		var r TopSourceIP
		if err := rows.Scan(&r.SourceIP, &r.Count, &r.Pass, &r.Fail); err != nil {
			return nil, fmt.Errorf("scan source IP row: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// DomainStats holds statistics for a single domain
type DomainStats struct {
	Domain            string  `json:"domain"`
	TotalMessages     int     `json:"total_messages"`
	CompliantMessages int     `json:"compliant_messages"`
	ComplianceRate    float64 `json:"compliance_rate"`
}

// OrgStats holds statistics for a reporting organization
type OrgStats struct {
	OrgName string `json:"org_name"`
	Reports int    `json:"reports"`
}

// DispositionStats holds statistics for a disposition type
type DispositionStats struct {
	Disposition string `json:"disposition"`
	Count       int    `json:"count"`
}

// AuthResultStats holds authentication result statistics
type AuthResultStats struct {
	Result string `json:"result"`
	Count  int    `json:"count"`
}

// GetDomainStats returns statistics grouped by domain
func (s *Storage) GetDomainStats() ([]DomainStats, error) {
	rows, err := s.db.Query(`
		SELECT domain,
		       COALESCE(SUM(total_messages), 0) as total_messages,
		       COALESCE(SUM(compliant_messages), 0) as compliant_messages
		FROM reports
		GROUP BY domain
	`)
	if err != nil {
		return nil, fmt.Errorf("query domain stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []DomainStats
	for rows.Next() {
		var ds DomainStats
		if err := rows.Scan(&ds.Domain, &ds.TotalMessages, &ds.CompliantMessages); err != nil {
			return nil, fmt.Errorf("scan domain stats row: %w", err)
		}
		if ds.TotalMessages > 0 {
			ds.ComplianceRate = float64(ds.CompliantMessages) / float64(ds.TotalMessages) * 100
		}
		stats = append(stats, ds)
	}
	return stats, nil
}

// GetOrgStats returns statistics grouped by reporting organization
func (s *Storage) GetOrgStats() ([]OrgStats, error) {
	rows, err := s.db.Query(`
		SELECT org_name, COUNT(*) as reports
		FROM reports
		GROUP BY org_name
	`)
	if err != nil {
		return nil, fmt.Errorf("query org stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []OrgStats
	for rows.Next() {
		var os OrgStats
		if err := rows.Scan(&os.OrgName, &os.Reports); err != nil {
			return nil, fmt.Errorf("scan org stats row: %w", err)
		}
		stats = append(stats, os)
	}
	return stats, nil
}

// GetDispositionStats returns message counts grouped by disposition
func (s *Storage) GetDispositionStats() ([]DispositionStats, error) {
	rows, err := s.db.Query(`
		SELECT COALESCE(disposition, 'unknown') as disposition,
		       SUM(count) as total_count
		FROM records
		GROUP BY disposition
	`)
	if err != nil {
		return nil, fmt.Errorf("query disposition stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []DispositionStats
	for rows.Next() {
		var ds DispositionStats
		if err := rows.Scan(&ds.Disposition, &ds.Count); err != nil {
			return nil, fmt.Errorf("scan disposition stats row: %w", err)
		}
		stats = append(stats, ds)
	}
	return stats, nil
}

// GetSPFStats returns SPF authentication result statistics
func (s *Storage) GetSPFStats() ([]AuthResultStats, error) {
	rows, err := s.db.Query(`
		SELECT COALESCE(spf_result, 'unknown') as result,
		       SUM(count) as total_count
		FROM records
		GROUP BY spf_result
	`)
	if err != nil {
		return nil, fmt.Errorf("query SPF stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []AuthResultStats
	for rows.Next() {
		var as AuthResultStats
		if err := rows.Scan(&as.Result, &as.Count); err != nil {
			return nil, fmt.Errorf("scan SPF stats row: %w", err)
		}
		stats = append(stats, as)
	}
	return stats, nil
}

// GetDKIMStats returns DKIM authentication result statistics
func (s *Storage) GetDKIMStats() ([]AuthResultStats, error) {
	rows, err := s.db.Query(`
		SELECT COALESCE(dkim_result, 'unknown') as result,
		       SUM(count) as total_count
		FROM records
		GROUP BY dkim_result
	`)
	if err != nil {
		return nil, fmt.Errorf("query DKIM stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []AuthResultStats
	for rows.Next() {
		var as AuthResultStats
		if err := rows.Scan(&as.Result, &as.Count); err != nil {
			return nil, fmt.Errorf("scan DKIM stats row: %w", err)
		}
		stats = append(stats, as)
	}
	return stats, nil
}
