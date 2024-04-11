/*
 * Copyright (C) 2024 Vladimir Homutov
 */

/*
 * This file is part of Rieman.
 *
 * Rieman is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Rieman is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 */

package main

import (
    "fmt"
    "os"
    "io"
    "log"
    "log/syslog"
)

type DaemonConfig struct {
    AppID             string
    LogFile           string
    Syslog            bool
    Stderr            bool
}

func init_daemon(cfg *DaemonConfig) error {

    var outputs []io.Writer

    log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix);
    log.SetPrefix(cfg.AppID + ": ");

    if cfg.Stderr {
        outputs = append(outputs, os.Stderr)
    }

    if cfg.Syslog {
        syslogger, err := syslog.New(syslog.LOG_INFO, cfg.AppID)
        if err != nil {
            return fmt.Errorf("failed to open syslog: %v\n", err)
        }

        outputs = append(outputs, syslogger)
    }

    if len(cfg.LogFile) != 0 {
        f, err := os.OpenFile(cfg.LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
        if err != nil {
            return fmt.Errorf("failed to open log file '%s': %v", cfg.LogFile, err)
        }

        outputs = append(outputs, f)
    }

    multi := io.MultiWriter(outputs...)
    log.SetOutput(multi)

    return nil
}
