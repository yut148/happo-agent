### 1.0.1 (Fri Feb 9 16:44:26 2018 +0900)

- Merge pull request #30 from heartbeatsjp/glide-to-dep (Wed Feb 7 16:44:44 2018 +0900) c8892ca
- Merge pull request #31 from heartbeatsjp/full-stacktrace (Fri Feb 9 11:47:22 2018 +0900) 74011b1
- Merge pull request #32 from heartbeatsjp/add-leveldb-properties-to-status (Fri Feb 9 16:40:35 2018 +0900) bc9def2

### 1.0.0 (Mon Sep 25 14:57:52 2017 +0900)

- Merge pull request #29 from heartbeatsjp/release-1.0.0pre4 (Mon Sep 25 13:09:38 2017 +0900) 74dd241
- Merge branch 'use-os-init' of github.com:heartbeatsjp/happo-agent into use-os-init (Thu Aug 10 18:44:56 2017 +0900) 3efebec
- Merge branch 'use-os-init' of github.com:heartbeatsjp/happo-agent into use-os-init (Thu Aug 10 18:53:43 2017 +0900) 775ebb5
- Merge branch 'use-os-init' of github.com:heartbeatsjp/happo-agent into use-os-init (Thu Sep 21 10:01:54 2017 +0900) 7b83b91
- Merge pull request #16 from heartbeatsjp/use-os-init (Mon Sep 25 12:38:47 2017 +0900) ca62134
- Merge pull request #17 from heartbeatsjp/happo-agent-add-return-error (Mon Jun 5 14:21:22 2017 +0900) 09226cd
- Merge pull request #18 from heartbeatsjp/use-dbms (Tue Jun 27 14:16:21 2017 +0900) 10808cb
- Merge pull request #19 from heartbeatsjp/passive-metrics (Mon Jul 3 12:42:13 2017 +0900) 01bf588
- Merge pull request #20 from heartbeatsjp/change-timeout-status-code (Thu Jul 13 18:23:27 2017 +0900) 5b2b1e4
- Merge pull request #21 from heartbeatsjp/fix-readme-ja (Wed Jul 26 10:02:34 2017 +0900) ffd225a
- Merge pull request #22 from heartbeatsjp/refactoring (Thu Aug 3 16:28:47 2017 +0900) 64bc7af
- Merge pull request #23 from heartbeatsjp/refactoring (Thu Aug 10 09:12:29 2017 +0900) 88cd7ab
- Merge pull request #24 from heartbeatsjp/enable-to-change-some-fixed-values (Tue Aug 15 17:14:05 2017 +0900) e0520f3
- Merge pull request #25 from heartbeatsjp/add-loglevel (Thu Aug 24 13:39:59 2017 +0900) 69b644b
- Merge pull request #26 from heartbeatsjp/improve-monitoring (Tue Sep 12 15:25:40 2017 +0900) 76440d1
- Merge pull request #27 from heartbeatsjp/re-design-response (Tue Sep 19 13:24:40 2017 +0900) ba394c4
- Merge pull request #28 from heartbeatsjp/re-design-response (Tue Sep 19 13:29:24 2017 +0900) 0a40b29

### 0.9.9 (Fri Sep 9 18:06:44 2016 +0900)

- POC for change daemon uid (Fri Jun 10 14:10:38 2016 +0900) 93e348f
- initial import (Wed Aug 3 15:07:53 2016 +0900) fc7f164
- add license description (Thu Aug 4 14:14:46 2016 +0900) 8e10644
- Merge pull request #12 from heartbeatsjp/license (Mon Aug 8 14:19:29 2016 +0900) 13d5149
- initial import (Thu Sep 8 20:07:20 2016 +0900) 8b4927f
- add badge (Thu Sep 8 20:08:47 2016 +0900) ef02afe
- Merge pull request #13 from heartbeatsjp/vendoring (Thu Sep 8 20:13:15 2016 +0900) bd23e13
- Change the status from MONITOR_UNKNOWN to MONITOR_ERROR when command timed out (Fri Sep 9 15:27:11 2016 +0900) 1365114
- Merge pull request #14 from rrreeeyyy/avoid-flapping-command-timed-out (Fri Sep 9 17:52:38 2016 +0900) b26c1f7

### 0.9.8 (Thu May 12 06:36:44 2016 +0900)

- Fixed #9 (Wed May 11 19:04:04 2016 +0900) 301c668
- Merge pull request #11 from heartbeatsjp/bugfix/lost-stderr-when-inventory-collection (Thu May 12 06:36:03 2016 +0900) 41606a2

### 0.9.7 (Tue Feb 16 01:17:30 2016 +0900)

- reduce log output when MARTINI_ENV=production (Tue Feb 16 01:13:10 2016 +0900) 6cb5ded
- Merge pull request #8 from heartbeatsjp/feature/reduce-log-in-production (Tue Feb 16 01:16:06 2016 +0900) 04846eb

### 0.9.6 (Mon Feb 15 23:32:35 2016 +0900)

- change martini.Logger to util.Logger (Mon Feb 15 16:55:45 2016 +0900) f49fb10
- add HUP handling (Mon Feb 15 19:51:10 2016 +0900) 73dee8c
- forget to commit. (Mon Feb 15 20:54:27 2016 +0900) 17eee99
- add logfile reopen when HUP (Mon Feb 15 20:57:14 2016 +0900) 754d0f6
- Merge pull request #7 from heartbeatsjp/feature/improve-logging (Mon Feb 15 23:32:13 2016 +0900) ab79aec

### 0.9.5 (Mon Feb 15 15:13:05 2016 +0900)

- unsophisticated implementation. (Fri Feb 12 15:24:00 2016 +0900) d45fc77
- add fast repeating restart guard. (Fri Feb 12 16:20:38 2016 +0900) 814d99e
- add signal propagation (Fri Feb 12 16:54:41 2016 +0900) 101ebb8
- Merge pull request #6 from heartbeatsjp/feature/subprocess-for-auto-restart (Mon Feb 15 15:12:05 2016 +0900) d7770f3

### 0.9.4 (Mon Feb 8 10:33:31 2016 +0900)

- - Implement metric_data_buffer capacity limitation - Add endpoint URI to get metric_data_buffer status (Fri Feb 5 16:03:21 2016 +0900) 576648c
- update README (Fri Feb 5 16:34:36 2016 +0900) d614e3a
- Merge pull request #5 from heartbeatsjp/feature/fuzzy_metrics_data_buffer_limit (Mon Feb 8 10:33:01 2016 +0900) fcd020b

### 0.9.3 (Thu Feb 4 13:14:32 2016 +0900)

- Version variable make UNinitialized. (Thu Feb 4 13:05:20 2016 +0900) afae2ba
- Merge pull request #4 from heartbeatsjp/feature/set_version_from_git_tag (Thu Feb 4 13:12:41 2016 +0900) d011cca

### 0.9.2 (Thu Feb 4 12:16:48 2016 +0900)

- Add lock to metrics_data_buffer (Wed Feb 3 15:20:33 2016 +0900) a9c100d
- - avoid race condition (closure is not groutine-safe) - saveMachineState() should not run synchronous.   synchronous execution leads to delay of monitor request (Wed Feb 3 18:43:59 2016 +0900) 8595c16
- Merge pull request #2 from heartbeatsjp/bugfix/lock_metrics_data_buffer (Thu Feb 4 12:16:37 2016 +0900) b8e4262
- Merge pull request #3 from heartbeatsjp/bugfix/safe_save_machine_state (Thu Feb 4 12:16:48 2016 +0900) e6d6d9a

### 0.9.1 (Mon Dec 14 16:00:50 2015 +0900)

- Can define port number at execute is_added and remove commands. (Mon Dec 14 15:50:35 2015 +0900) 89dea60
- Bump up version (Mon Dec 14 15:59:36 2015 +0900) 6b2dde7

