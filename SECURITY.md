# Security

`go-exec-format-doctor` treats every inspected file as untrusted data. It reads
at most the first 4 KiB and does not execute, modify, upload, or make network
requests with the file.

Do not use its compatibility verdict as a malware or safety assessment.

For a suspected vulnerability in this repository, open a GitHub security
advisory instead of publishing exploit details in a public issue.
