(SELECT uuid,
        source_code,
        status,
        stats,
        trigger,
        error_details,
        started_at,
        completed_at
 FROM vulnerability_database_updates
 WHERE source_code = 'nvd'
   AND status IN ('completed', 'failed')
 ORDER BY started_at DESC
     LIMIT 1)
UNION
(SELECT uuid,
        source_code,
        status,
        stats,
        trigger,
        error_details,
        started_at,
        completed_at
 FROM vulnerability_database_updates
 WHERE source_code = 'osv'
   AND status IN ('completed', 'failed')
 ORDER BY started_at DESC
     LIMIT 1)
UNION
(SELECT uuid,
        source_code,
        status,
        stats,
        trigger,
        error_details,
        started_at,
        completed_at
 FROM vulnerability_database_updates
 WHERE source_code = 'ghsa'
   AND status IN ('completed', 'failed')
 ORDER BY started_at DESC
     LIMIT 1)
UNION
(SELECT uuid,
        source_code,
        status,
        stats,
        trigger,
        error_details,
        started_at,
        completed_at
 FROM vulnerability_database_updates
 WHERE source_code = 'kev'
   AND status IN ('completed', 'failed')
 ORDER BY started_at DESC
     LIMIT 1)
UNION
(SELECT uuid,
        source_code,
        status,
        stats,
        trigger,
        error_details,
        started_at,
        completed_at
 FROM vulnerability_database_updates
 WHERE source_code = 'epss'
   AND status IN ('completed', 'failed')
 ORDER BY started_at DESC
     LIMIT 1)