// SPDX-License-Identifier: GPL-3.0-or-later

package mysql

import "regexp"

const (
	queryShowGlobalVariables = `
SHOW GLOBAL VARIABLES 
WHERE 
  Variable_name LIKE 'max_connections' 
  OR Variable_name LIKE 'table_open_cache' 
  OR Variable_name LIKE 'disabled_storage_engines' 
  OR Variable_name LIKE 'log_bin'
  OR Variable_name LIKE 'innodb_log_file_size'
  OR Variable_name LIKE 'innodb_log_files_in_group'
  OR Variable_name LIKE 'wsrep_provider_options'
  OR Variable_name LIKE 'performance_schema';`
)

var reGCacheKeepPagesSize = regexp.MustCompile(`gcache\.keep_pages_size\s*=\s*(\d+)([KMGT]+);`)

func (m *MySQL) collectGlobalVariables() error {
	// MariaDB: https://mariadb.com/kb/en/server-system-variables/
	// MySQL: https://dev.mysql.com/doc/refman/8.0/en/server-system-variable-reference.html
	q := queryShowGlobalVariables
	m.Debugf("executing query: '%s'", q)

	var name string
	_, err := m.collectQuery(q, func(column, value string, _ bool) {
		switch column {
		case "Variable_name":
			name = value
		case "Value":
			switch name {
			case "disabled_storage_engines":
				m.varDisabledStorageEngine = value
			case "innodb_log_file_size":
				m.varInnodbLogFileSize = parseInt(value)
			case "innodb_log_files_in_group":
				m.varInnodbLogFilesInGroup = parseInt(value)
			case "log_bin":
				m.varLogBin = value
			case "max_connections":
				m.varMaxConns = parseInt(value)
			case "performance_schema":
				m.varPerformanceSchema = value
			case "table_open_cache":
				m.varTableOpenCache = parseInt(value)
			case "wsrep_provider_options":
				match := reGCacheKeepPagesSize.FindStringSubmatch(value)
				if len(match) >= 2 {
					m.hasGCache = true
					val := parseInt(match[1])
					if len(match) == 3 {
						switch match[2] {
						case "K":
							val *= 1024
						case "M":
							val *= 1024 * 1024
						case "G":
							val *= 1024 * 1024 * 1024
						case "T":
							val *= 1024 * 1024 * 1024 * 1024
						}
					}
					m.varGCacheKeepPagesSize = val
				}
			}
		}
	})
	return err
}
