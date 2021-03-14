#!/bin/bash

#[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
conn="mysql://impart_super:supersecretpassword@tcp(localhost:3306)/impart?tls=skip-verify&autocommit=true"
migrate -database "${conn}" -path ./schemas/migrations "${1}"

#$ migrate -help
#Usage: migrate OPTIONS COMMAND [arg...]
#       migrate [ -version | -help ]
#
#Options:
#  -source          Location of the migrations (driver://url)
#  -path            Shorthand for -source=file://path
#  -database        Run migrations against this database (driver://url)
#  -prefetch N      Number of migrations to load in advance before executing (default 10)
#  -lock-timeout N  Allow N seconds to acquire database lock (default 15)
#  -verbose         Print verbose logging
#  -version         Print version
#  -help            Print usage
#
#Commands:
#  create [-ext E] [-dir D] [-seq] [-digits N] [-format] NAME
#               Create a set of timestamped up/down migrations titled NAME, in directory D with extension E.
#               Use -seq option to generate sequential up/down migrations with N digits.
#               Use -format option to specify a Go time format string.
#  goto V       Migrate to version V
#  up [N]       Apply all or N up migrations
#  down [N]     Apply all or N down migrations
#  drop         Drop everything inside database
#  force V      Set version V but don't run migration (ignores dirty state)
#  version      Print current migration version