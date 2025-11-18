#!/usr/bin/env python3
"""
Generic Process and Thread Monitoring Script
Monitors CPU and memory usage for all processes and threads in the container.
Outputs structured JSON logs that can be consumed by CloudWatch Logs.
"""

import os
import sys
import time
import json
import signal
from datetime import datetime
from typing import Dict, List, Any

try:
    import psutil
except ImportError:
    print("ERROR: psutil library not found. Install with: pip install psutil", file=sys.stderr)
    sys.exit(1)


class ProcessMonitor:
    def __init__(self, interval: int = 30, filter_process: str = None):
        """
        Initialize the process monitor.
        
        Args:
            interval: Collection interval in seconds (default: 30)
            filter_process: Optional process name filter (e.g., "serv" to only monitor processes containing "serv")
        """
        self.interval = interval
        self.filter_process = filter_process
        self.running = True
        
        # Register signal handlers for graceful shutdown
        signal.signal(signal.SIGTERM, self._signal_handler)
        signal.signal(signal.SIGINT, self._signal_handler)
    
    def _signal_handler(self, signum, frame):
        """Handle shutdown signals gracefully."""
        self.running = False
    
    def get_process_metrics(self, proc: psutil.Process) -> Dict[str, Any]:
        """
        Collect metrics for a single process.
        
        Args:
            proc: psutil.Process object
            
        Returns:
            Dictionary with process metrics
        """
        try:
            # Get process info
            proc_info = {
                "pid": proc.pid,
                "name": proc.name(),
                "exe": proc.exe() if proc.exe() else "N/A",
                "status": proc.status(),
                "cpu_percent": proc.cpu_percent(interval=0.1),
                "memory_rss_mb": proc.memory_info().rss / 1024 / 1024,  # Convert to MB
                "memory_vms_mb": proc.memory_info().vms / 1024 / 1024,  # Convert to MB
                "memory_percent": proc.memory_percent(),
                "num_threads": proc.num_threads(),
                "create_time": proc.create_time(),
            }
            
            # Get thread-level metrics
            threads = []
            try:
                for thread in proc.threads():
                    thread_info = {
                        "tid": thread.id,
                        "cpu_percent": proc.cpu_percent(interval=0.1),  # Per-thread CPU is approximate
                    }
                    threads.append(thread_info)
            except (psutil.NoSuchProcess, psutil.AccessDenied):
                pass
            
            proc_info["threads"] = threads
            
            # Get additional info if available
            try:
                proc_info["cmdline"] = " ".join(proc.cmdline()) if proc.cmdline() else "N/A"
            except (psutil.NoSuchProcess, psutil.AccessDenied):
                proc_info["cmdline"] = "N/A"
            
            return proc_info
            
        except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess) as e:
            return {"pid": proc.pid, "error": str(e)}
    
    def collect_metrics(self) -> List[Dict[str, Any]]:
        """
        Collect metrics for all processes.
        
        Returns:
            List of process metric dictionaries
        """
        processes = []
        
        try:
            # Get all processes
            for proc in psutil.process_iter(['pid', 'name']):
                try:
                    # Apply filter if specified
                    if self.filter_process:
                        proc_name = proc.info.get('name', '')
                        if self.filter_process.lower() not in proc_name.lower():
                            continue
                    
                    # Collect metrics
                    metrics = self.get_process_metrics(proc)
                    if metrics:
                        processes.append(metrics)
                        
                except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
                    # Process may have terminated, skip it
                    continue
                    
        except Exception as e:
            print(json.dumps({
                "timestamp": datetime.utcnow().isoformat(),
                "level": "ERROR",
                "message": f"Error collecting process metrics: {str(e)}"
            }), file=sys.stderr)
        
        return processes
    
    def run(self):
        """Main monitoring loop."""
        print(json.dumps({
            "timestamp": datetime.utcnow().isoformat(),
            "level": "INFO",
            "message": f"Process monitor started (interval: {self.interval}s, filter: {self.filter_process or 'all'})"
        }), flush=True)
        
        while self.running:
            try:
                # Collect metrics
                processes = self.collect_metrics()
                
                # Output metrics as JSON
                output = {
                    "timestamp": datetime.utcnow().isoformat(),
                    "type": "process_metrics",
                    "total_processes": len(processes),
                    "processes": processes
                }
                
                print(json.dumps(output), flush=True)
                
                # Wait for next interval
                time.sleep(self.interval)
                
            except KeyboardInterrupt:
                self.running = False
                break
            except Exception as e:
                print(json.dumps({
                    "timestamp": datetime.utcnow().isoformat(),
                    "level": "ERROR",
                    "message": f"Error in monitoring loop: {str(e)}"
                }), file=sys.stderr, flush=True)
                time.sleep(self.interval)
        
        print(json.dumps({
            "timestamp": datetime.utcnow().isoformat(),
            "level": "INFO",
            "message": "Process monitor stopped"
        }), flush=True)


def main():
    """Main entry point."""
    # Get configuration from environment variables
    interval = int(os.getenv("MONITOR_INTERVAL", "30"))
    filter_process = os.getenv("MONITOR_FILTER", None)
    
    # Create and run monitor
    monitor = ProcessMonitor(interval=interval, filter_process=filter_process)
    monitor.run()


if __name__ == "__main__":
    main()

