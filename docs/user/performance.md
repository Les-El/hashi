# Performance Tuning & Resource Management

Chexum is designed to be a "good neighbor" on your system. Unlike many CLI tools that greedily consume every available CPU cycle, Chexum follows a strict **Neighborhood Policy** to ensure your computer remains responsive, even while hashing terabytes of data.

## Default Behavior (Auto-Pilot)

When you run `chexum` without any special flags, it automatically detects your system's hardware and sets a safe concurrency limit.

| System Core Count | Chexum Default | Protocol |
|-------------------|---------------|----------|
| 1 core (VMs) | 1 worker | Single-threaded safety |
| 2-4 cores | N-1 workers | Leave 1 core for OS |
| >4 cores | N-2 workers | Leave 2 cores for OS/Apps |
| Massive Servers | Max 32 workers | Prevent context-switching storms |

This ensures that you can run a massive hash operation in the background while still browsing the web, watching videos, or compiling code without the "micro-stuttering" often caused by 100% CPU load.

## Manual Override (`--jobs`)

If you want to unlock full power (e.g., on a dedicated CI/CD server) or restrict Chexum further (e.g., background cron job), you can use the `--jobs` (or `-j`) flag.

### Examples

**Max Power (Not recommended for desktops):**
```bash
# Force 100 threads
chexum --jobs 100 ./massive-iso-collection
```
*Note: This will likely slow down your system due to scheduler overhead.*

**Background Mode (Silent & Light):**
```bash
# Use exactly 1 core
chexum --jobs 1 --quiet ./backups
```

## OS Integration (cgroups & Docker)

Chexum respects the Go runtime's awareness of OS-level constraints.
-   **Docker/Kubernetes**: If you limit a container to 2 CPUs, `chexum` detects 2 CPUs and scales accordingly.
-   **Affinity**: If you use `taskset` to bind `chexum` to specific cores, it will only use those cores.
