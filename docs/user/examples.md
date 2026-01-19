# hashi - User Examples Guide

This document provides detailed examples for every capability of the `hashi` tool, demonstrating how to use its features to solve real-world problems.

---

## 1. Basic Usage

### 1.1 Hash Current Directory
**Scenario:** You want to quickly check the hashes of all files in your current working folder.
**Command:**
```bash
hashi
```
**Output:**
```text
[SHA-256] Computed hashes for 3 files:
e3b0c442...  notes.txt
a1b2c3d4...  image.png
88d4266f...  data.csv
```
**Explanation:** Running `hashi` without arguments processes all visible files in the current directory using the default SHA-256 algorithm.

### 1.2 Hash a Single File
**Scenario:** You need the hash of a specific ISO file to share with a friend.
**Command:**
```bash
hashi ubuntu-24.04.iso
```
**Output:**
```text
[SHA-256] ubuntu-24.04.iso
9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```
**Explanation:** `hashi` calculates the SHA-256 hash for the specified file and displays it clearly.

### 1.3 Verify a File Integrity (Pass)
**Scenario:** You downloaded a file and want to verify it matches the hash provided by the website.
**Command:**
```bash
hashi installer.zip a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2
```
**Output:**
```text
[SHA-256] Verifying installer.zip...
✅ PASS: Hash matches provided string.
```
**Explanation:** You provided both a filename and a hash string. `hashi` automatically detected this, computed the file's hash, compared it to your string, and gave a simple "PASS" result.

### 1.4 Verify a File Integrity (Fail)
**Scenario:** Same as above, but the file was corrupted during download.
**Command:**
```bash
hashi installer.zip a1b2c3d4e5f6... (expected hash)
```
**Output:**
```text
[SHA-256] Verifying installer.zip...
🔴 FAIL: Hash mismatch!
   Expected: a1b2c3d4e5f6...
   Computed: f0e1d2c3b4a5...
```
**Explanation:** `hashi` detected that the computed hash did not match your input, alerting you to potential corruption or tampering.

### 1.5 Validate a Hash String
**Scenario:** You have a hash string and want to check if it's a valid format.
**Command:**
```bash
hashi a1b2c3d4...
```
**Output:**
```text
✅ Valid SHA-256 hash format.
```
**Explanation:** Since no file matched the argument, `hashi` checked if it was a valid hash string format.

---

## 2. File Selection & Traversal

### 2.1 Recursive Hashing (`-a`)
**Scenario:** You want to hash every file in your project, including those in subfolders `src/` and `images/`.
**Command:**
```bash
hashi -a
```
**Output:**
```text
[SHA-256] Computed hashes for 15 files (recursive):
...
7d793037...  src/main.go
b1946ac9...  src/utils/helper.go
55502f40...  images/logo.png
...
```
**Explanation:** The `-a` (all subdirectories) flag tells `hashi` to traverse the directory tree downwards.

### 2.2 Include Hidden Files (`-H`)
**Scenario:** You need to check configuration files like `.bashrc` or `.git/config`.
**Command:**
```bash
hashi -H
```
**Output:**
```text
[SHA-256] Computed hashes for 5 files (including hidden):
...
324d26c5...  .gitignore
982d9212...  .env
...
```
**Explanation:** The `-H` flag forces `hashi` to include files and directories starting with a dot (`.`), which are usually ignored.

### 2.3 Recursive + Hidden (`-a -H`)
**Scenario:** You want a complete manifest of everything in the folder, hidden or not, deep or shallow.
**Command:**
```bash
hashi -a -H
```
**Output:**
```text
[SHA-256] Computed hashes for 42 files (recursive, hidden):
...
12345678...  .git/HEAD
87654321...  src/.temp_cache
...
```
**Explanation:** Combining flags works intuitively. `hashi` crawls the entire tree including hidden items.

---

## 3. Filtering

### 3.1 Include by Pattern (`--include`)
**Scenario:** You only care about your source code files.
**Command:**
```bash
hashi -a --include "*.go,*.js"
```
**Output:**
```text
[SHA-256] Filtering: include=[*.go, *.js]
Computed hashes for 12 files:
...
```
**Explanation:** `hashi` scanned the tree but only processed files ending in `.go` or `.js`.

### 3.2 Exclude by Pattern (`--exclude`)
**Scenario:** You want to hash everything except temporary log files.
**Command:**
```bash
hashi -a --exclude "*.log"
```
**Output:**
```text
[SHA-256] Filtering: exclude=[*.log]
Computed hashes for 8 files:
...
```
**Explanation:** Files matching `*.log` were skipped.

### 3.3 Filter by Size (`--min-size`, `--max-size`)
**Scenario:** You want to find large ISOs (>1GB) but ignore small text files.
**Command:**
```bash
hashi -a --min-size 1GB
```
**Output:**
```text
[SHA-256] Filtering: size >= 1GB
Computed hashes for 2 files:
e3b0c442...  backups/full_db.dump
a1b2c3d4...  downloads/movie.mkv
```
**Explanation:** Only files meeting the size criteria were hashed.

### 3.4 Filter by Date (`--modified-after`)
**Scenario:** You verified your backup last week. You only want to check files changed since then.
**Command:**
```bash
hashi -a --modified-after 2026-01-10
```
**Output:**
```text
[SHA-256] Filtering: modified >= 2026-01-10
Computed hashes for 4 files:
...
```
**Explanation:** `hashi` checked file metadata and only processed recently modified files.

---

## 4. Advanced Features

### 4.1 Auto-Algorithm Detection
**Scenario:** A website gives you a short 32-character hash (MD5), but you forget to specify `--algo md5`.
**Command:**
```bash
hashi myfile.exe 5d41402abc4b2a76b9719d911017c592
```
**Output:**
```text
[INFO] Auto-detected algorithm: MD5 (based on 32-char length)
[MD5] Verifying myfile.exe...
✅ PASS: Hash matches.
```
**Explanation:** `hashi` noticed the string length was 32, inferred it was likely MD5, switched the algorithm for you, and verified the file.

### 4.2 Directory Comparison (`--compare`)
**Scenario:** You copied a folder to a backup drive and want to ensure it's identical.
**Command:**
```bash
hashi --compare ./photos /mnt/backup/photos
```
**Output:**
```text
Directory Comparison Results:
-----------------------------
✅ 150 files match perfectly.

🔴 MISMATCHES (Content differs):
   - DCIM/IMG_001.jpg

🟡 UNIQUE to ./photos:
   - new_pic.jpg

🟡 UNIQUE to /mnt/backup/photos:
   - old_deleted_pic.jpg
```
**Explanation:** `hashi` diffed the two directory trees, identifying content mismatches (corruption), and missing/extra files.

### 4.3 Explicit Algorithm Selection (`--algo`)
**Scenario:** You need a SHA-512 hash specifically.
**Command:**
```bash
hashi --algo sha512 secure_doc.pdf
```
**Output:**
```text
[SHA-512] secure_doc.pdf
cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce... (truncated)
```
**Explanation:** The `--algo` flag overrides the default SHA-256.

---

## 5. Output & Logging

### 5.1 JSON Output (`--output-format json`)
**Scenario:** You are writing a Python script to process these hashes and need structured data.
**Command:**
```bash
hashi -a --output-format json
```
**Output:**
```json
[
  {
    "file": "notes.txt",
    "hash": "e3b0c442...",
    "algorithm": "sha256",
    "size": 1024
  },
  ...
]
```
**Explanation:** The output is pure JSON, ready to be piped into `jq` or read by other programs.

### 5.2 Logging to File (`--log-file`)
**Scenario:** You are running a long verification overnight and want a record of any errors.
**Command:**
```bash
hashi -a --compare /data /backup --log-file verify.log
```
**Output:**
```text
(Screen shows progress bar...)
```
**File Content (verify.log):**
```text
2026-01-16 10:00:01 [INFO] Starting comparison of /data and /backup
2026-01-16 10:05:23 [ERROR] Read error on /data/corrupt_sector.bin: input/output error
...
```
**Explanation:** The screen remains clean (or shows progress), while detailed event logs are written to disk.

### 5.3 Save Output Report (`--output-file`)
**Scenario:** You want to save the "PASS/FAIL" summary report to a text file for your records.
**Command:**
```bash
hashi file.zip <hash> --output-file report.txt
```
**Output:**
```text
(No output to screen)
```
**File Content (report.txt):**
```text
[SHA-256] Verifying file.zip...
✅ PASS: Hash matches provided string.
```
**Explanation:** The human-readable report is redirected to the file.

---

## 6. Configuration Files

### 6.1 Using a Config File (`--config`)
**Scenario:** You have a complex set of exclusions and folders you check every day.
**Command:**
```bash
hashi --config daily_check.json
```
**Content of daily_check.json:**
```json
{
  "recursive": true,
  "algo": "md5",
  "exclude": ["*.tmp", "*.cache"],
  "paths": ["./project", "./docs"]
}
```
**Output:**
```text
[MD5] Loading configuration from daily_check.json...
Computed hashes for 150 files...
```
**Explanation:** `hashi` reads the flags and arguments from the file, saving you from typing them every time.
