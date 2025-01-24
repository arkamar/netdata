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
  OR Variable_name LIKE 'wsrep_provider_options'
  OR Variable_name LIKE 'performance_schema';`
)

var reGCacheKeepPagesSize = regexp.MustCompile(`gcache\.keep_pages_size\s*=\s*(\d+)([KMGT]+);`)

func (c *Collector) collectGlobalVariables() error {
	// MariaDB: https://mariadb.com/kb/en/server-system-variables/
	// MySQL: https://dev.mysql.com/doc/refman/8.0/en/server-system-variable-reference.html
	q := queryShowGlobalVariables
	c.Debugf("executing query: '%s'", q)

	var name string
	_, err := c.collectQuery(q, func(column, value string, _ bool) {
		switch column {
		case "Variable_name":
			name = value
		case "Value":
			switch name {
			case "disabled_storage_engines":
				c.varDisabledStorageEngine = value
			case "innodb_log_file_size":
				c.varInnodbLogFileSize = parseInt(value)
			case "log_bin":
				c.varLogBin = value
			case "max_connections":
				c.varMaxConns = parseInt(value)
			case "performance_schema":
				c.varPerformanceSchema = value
			case "table_open_cache":
				c.varTableOpenCache = parseInt(value)
			case "wsrep_provider_options":
				match := reGCacheKeepPagesSize.FindStringSubmatch(value)
				if len(match) >= 2 {
					c.hasGCache = true
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
					c.varGCacheKeepPagesSize = val
				}
			}
		}
	})
	return err
}
